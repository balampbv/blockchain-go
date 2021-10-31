package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)

var bcServer chan []Block

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	bcServer = make(chan []Block)

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

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)
	}

}

//handleConn does the following repeatedly
// 1. Prompts the client to enter the bpm input
// 2. Scan the client input from stdin
// 3. Create a new block with this input data, using generateBlock, isValidBlock and replaceChain funcs
// 4. Put the new blockchain in the channel we created to broadcast to the network
// 5. Allow the client to enter a new bpm input
func handleConn(conn net.Conn) {
	defer conn.Close()

	io.WriteString(conn, "Enter a new BPM: ")

	scanner := bufio.NewScanner(conn)

	go func() {
		for scanner.Scan() {

			bpm, err := strconv.Atoi(scanner.Text())
			if err != nil {
				log.Printf("%v not a number: %v", scanner.Text(), err)
				continue
			}
			newBlock, err := generateBlock(BlockChain[len(BlockChain)-1], bpm)
			if err != nil {
				log.Println(err)
				continue
			}

			if isValidBlock(newBlock, BlockChain[len(BlockChain)-1]) {
				newBlockChain := append(BlockChain, newBlock)
				replaceChain(newBlockChain)
			}

			bcServer <- BlockChain

			io.WriteString(conn, "\nEnter a new BPM: ")
		}
	}()

	//simulate receiving broadcast
	go func() {
		for {
			time.Sleep(30 * time.Second)
			output, err := json.Marshal(BlockChain)
			if err != nil {
				log.Fatal(err)
			}
			io.WriteString(conn, string(output))
		}
	}()

	for _ = range bcServer {
		spew.Dump(BlockChain)
	}
}
