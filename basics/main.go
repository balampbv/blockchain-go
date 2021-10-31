package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Message struct {
	BPM int
}

//handleGetBlockchain gets the blockchain data and writes in the response writer
func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(BlockChain, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

//handleWriteBlock decodes the body, generates a new block and validates it
func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var m Message

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

	newBlock, err := generateBlock(BlockChain[len(BlockChain)-1], m.BPM)
	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, m)
		return
	}
	if isValidBlock(newBlock, BlockChain[len(BlockChain)-1]) {
		newBlockchain := append(BlockChain, newBlock)
		replaceChain(newBlockchain)
		spew.Dump(BlockChain)
	}

	respondWithJSON(w, r, http.StatusCreated, newBlock)

}

//respondWithJSON writes the code and response body in the writer.
func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

//makeMuxRouter creates a router
func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}

//run creates the http server
func run() error {
	mux := makeMuxRouter()
	httpAddr := os.Getenv("PORT")
	log.Println("Listening on ", os.Getenv("PORT"))
	s := &http.Server{
		Addr:           "localhost:" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	//it creates the genesisblock for the first time
	go func() {
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
	}()

	log.Fatal(run())
}
