package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// True SO coder detected xD
// https://stackoverflow.com/a/22892986/1976993
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

var conn net.Conn

func randomizerHTTPServer(w http.ResponseWriter, r *http.Request) {
	requestedAmountString := r.URL.Path[1:]
	requestedAmount, err := strconv.Atoi(requestedAmountString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	randomStringList := make([]string, requestedAmount)
	for i := range randomStringList {
		randomStringList[i] = randSeq(10 + rand.Intn(20))
	}

	// Marshal json & send ws message
	jsonRequestBlob, _ := json.Marshal(randomStringList)
	err = wsutil.WriteClientMessage(conn, ws.OpText, jsonRequestBlob)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// wait encoded response
	encryptorResponse, _, err := wsutil.ReadServerData(conn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "%s", string(encryptorResponse))
}

func main() {
	rand.Seed(time.Now().UnixNano())
	svcPort := os.Getenv("SERVICE_PORT")
	if svcPort == "" {
		svcPort = "6000"
	}
	encEndpoint := os.Getenv("ENCRYPTOR_ENDPOINT")
	if encEndpoint == "" {
		encEndpoint = "ws://localhost:5000/"
	}
	// init
	conn1, _, _, err := ws.DefaultDialer.Dial(context.Background(), encEndpoint)
	conn = conn1
	if err != nil {
		fmt.Printf("Cannot connect to Encryptor microservice at %s.\n", encEndpoint)
		return
	}

	fmt.Printf("StringRandomizer microservice is listening on port %s.\n", svcPort)
	http.HandleFunc("/", randomizerHTTPServer)
	http.ListenAndServe(":"+svcPort, nil)
}
