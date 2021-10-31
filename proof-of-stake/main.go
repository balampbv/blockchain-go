package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	//Create the genesis block
	t := time.Now()
	genesisBlock := Block{
		Index:     0,
		Timestamp: t.String(),
		BPM:       0,
		Hash:      "",
		PrevHash:  "",
	}
	spew.Dump(genesisBlock)
	BlockChain = append(BlockChain, genesisBlock)

	//Start TCP and serve TCP server
	server, err := net.Listen("tcp", ":"+os.Getenv("PORT"))
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	go func() {
		for candidate := range candidateBlocks {
			mutex.Lock()
			tempBlocks = append(tempBlocks, candidate)
			mutex.Unlock()
		}
	}()

	go func() {
		for {
			pickWinner()
		}
	}()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)
	}

}

func handleConn(conn net.Conn) {
	defer conn.Close()

	go func() {
		for {
			msg := <-announcements
			io.WriteString(conn, msg)
		}
	}()

	// Validator address
	var address string
	// allow user to allocate number of tokens to stake
	// greate the number of tokens, the greater chance to forging a new block.
	io.WriteString(conn, "Enter token balance: ")
	scanBalance := bufio.NewScanner(conn)

	for scanBalance.Scan() {
		balance, err := strconv.Atoi(scanBalance.Text())
		if err != nil {
			log.Printf("%v not a number: %v", scanBalance.Text(), err)
			return
		}
		t := time.Now()

		address = calculateHash(t.String())
		validators[address] = balance
		fmt.Println(validators)
		break
	}

	io.WriteString(conn, "Enter a new BPM: ")

	scanner := bufio.NewScanner(conn)

	go func() {
		for {
			//take in BPM from stdin and add it to blockchain after conducting necessary validations
			for scanner.Scan() {

				bpm, err := strconv.Atoi(scanner.Text())
				//if malicious party tries to mutate the chain with a bad input, delete them as a validator and they lose their staked tokens
				if err != nil {
					log.Printf("%v not a number: %v", scanner.Text(), err)
					delete(validators, address)
					conn.Close()
				}

				mutex.Lock()
				oldLastIndex := BlockChain[len(BlockChain)-1]
				mutex.Unlock()

				newBlock, err := generateBlock(oldLastIndex, bpm, address)
				if err != nil {
					log.Println(err)
					continue
				}

				if isValidBlock(newBlock, oldLastIndex) {
					candidateBlocks <- newBlock
				}

				io.WriteString(conn, "\nEnter a new BPM: ")
			}
		}
	}()

	//simulate receiving broadcast
	for {
		time.Sleep(60 * time.Second)
		output, err := json.Marshal(BlockChain)
		if err != nil {
			log.Fatal(err)
		}
		io.WriteString(conn, string(output))
	}
}

// pickWinner creates lotttery pool of validators and chooses the validator who gets to forge a block to the blockchain
// by random selecting from the pool, weighted by amount of tokens staked
func pickWinner() {
	time.Sleep(30 * time.Second)
	mutex.Lock()
	temp := tempBlocks
	mutex.Unlock()

	lotteryPool := []string{}

	if len(temp) > 0 {

		// slightly modified traditional proof of stake algorithm
		// from all validators who submitted a block, weight them by the number of staked tokens
		// in traditional proof of stake, validators can participate without submitting a block to be forged

	OUTER:

		for _, block := range temp {
			//if already in lottery pool, skip
			for _, node := range lotteryPool {
				if block.Validator == node {
					continue OUTER
				}
			}

			//lock the list of validators to prevent data race
			mutex.Lock()
			setValidators := validators
			mutex.Unlock()

			k, ok := setValidators[block.Validator]
			if ok {
				for i := 0; i < k; i++ {
					lotteryPool = append(lotteryPool, block.Validator)
				}
			}

		}

		// randomly pick winner from lottery pool
		s := rand.NewSource(time.Now().Unix())
		r := rand.New(s)
		lotteryWinner := lotteryPool[r.Intn(len(lotteryPool))]

		// add block of winner to blockchain and let all the other nodes know
		for _, block := range temp {
			if block.Validator == lotteryWinner {
				mutex.Lock()
				BlockChain = append(BlockChain, block)
				mutex.Unlock()
				for _ = range validators {
					announcements <- "\nwinning validator: " + lotteryWinner + "\n"
				}
				break
			}
		}
	}

	mutex.Lock()
	tempBlocks = []Block{}
	mutex.Unlock()
}
