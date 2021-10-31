package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type Block struct {
	Index      int
	Timestamp  string
	BPM        int
	Hash       string
	PrevHash   string
	Difficulty int
	Nounce     string
}

//BlockChain holds the list of blocks
var BlockChain []Block

//calculateHash to calcluate the hash value based on the block inputs
func calculateHash(block Block) string {

	record := string(rune(block.Index)) + block.Timestamp + string(rune(block.BPM)) + block.PrevHash + block.Nounce
	h := sha256.New()
	h.Write([]byte(record))
	hased := h.Sum(nil)
	// fmt.Println("hased :", hased)
	// fmt.Println("hex.EncodeTosString :", hex.EncodeToString(hased))

	return hex.EncodeToString(hased)

}

//generateBlock takes previoublock and new bpm values as input and creates a new block out of it
func generateBlock(previousBlock Block, BPM int) (Block, error) {

	newBlock := Block{
		Index:      previousBlock.Index + 1,
		Timestamp:  time.Now().String(),
		BPM:        BPM,
		PrevHash:   previousBlock.Hash,
		Difficulty: 1, //Set to 1 for simplicity
	}
	// newBlock.Hash = calculateHash(newBlock)

	for i := 0; ; i++ {
		hex := fmt.Sprintf("%x", i)
		newBlock.Nounce = hex
		if !isHashValid(calculateHash(newBlock), newBlock.Difficulty) {
			fmt.Println(calculateHash(newBlock), " do some more work!")
			time.Sleep(time.Second)
			continue
		} else {
			fmt.Println(calculateHash(newBlock), "Work done!")
			newBlock.Hash = calculateHash(newBlock)
			break
		}

	}

	return newBlock, nil
}

//isValidBlock takes new and old block as argments and validates the integrity of the blocks
func isValidBlock(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

//replaceChain is a tiebreaker function to resolve which chain to pick up when multiple nodes generates the block
//and appends in the chain. Two well meaning nodes may simply have different chain lengths,
//so naturally the longer one will be the most up to date and have the latest blocks.
func replaceChain(newChain []Block) {
	if len(newChain) > len(BlockChain) {
		BlockChain = newChain
	}
}

//isHashValid checks whether the hash generated matches the proof-of-work requirements
//checks if the number of leading zeros defined in the difficulty matches the hash generated
func isHashValid(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}
