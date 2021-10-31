package main

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
	Validator string
}

//BlockChain holds the list of blocks
var BlockChain []Block
var tempBlocks []Block

// candidateBlocks handles incoming blocks for validation
var candidateBlocks = make(chan Block)

// announcements broadcasts winning validator to all nodes
var announcements = make(chan string)

var mutex = &sync.Mutex{}

// validators keeps track of open validators and balances
var validators = make(map[string]int)

// SHA256 hasing
// calculateHash is a simple SHA256 hashing function
func calculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

//calculateBlockHash returns the hash of all block information
func calculateBlockHash(block Block) string {
	record := string(rune(block.Index)) + block.Timestamp + string(rune(block.BPM)) + block.PrevHash
	return calculateHash(record)
}

//generateBlock takes previoublock and new bpm values as input and creates a new block out of it
func generateBlock(previousBlock Block, BPM int, address string) (Block, error) {

	newBlock := Block{
		Index:     previousBlock.Index + 1,
		Timestamp: time.Now().String(),
		BPM:       BPM,
		PrevHash:  previousBlock.Hash,
	}
	newBlock.Hash = calculateBlockHash(newBlock)
	newBlock.Validator = address

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

	if calculateBlockHash(newBlock) != newBlock.Hash {
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
