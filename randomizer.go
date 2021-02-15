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

var (
	conn                 net.Conn
	minLength, maxLength int
)

func randomizerHTTPServer(w http.ResponseWriter, r *http.Request) {
	requestedAmountString := r.URL.Path[1:]
	requestedAmount, err := strconv.Atoi(requestedAmountString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	randomStringList := make([]string, requestedAmount)
	for i := range randomStringList {
		randomStringList[i] = randSeq(minLength + rand.Intn(maxLength-minLength))
	}

	// Marshal json & send ws message
	jsonRequestBlob, _ := json.Marshal(randomStringList)
	err = wsutil.WriteClientMessage(conn, ws.OpText, jsonRequestBlob)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Wait encoded response
	encryptorResponse, _, err := wsutil.ReadServerData(conn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Serve response, accelerate by manually making valid json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{\"Input\":%s,\"Output\":%s}", jsonRequestBlob, encryptorResponse)
	//w.Write(encryptorResponse)
}

func main() {
	var err error

	rand.Seed(time.Now().UnixNano())

	svcPort := os.Getenv("SERVICE_PORT")
	if svcPort == "" {
		svcPort = "6000"
	}

	encEndpoint := os.Getenv("ENCRYPTOR_ENDPOINT")
	if encEndpoint == "" {
		encEndpoint = "ws://localhost:5000/"
	}

	minLengthStr := os.Getenv("STRING_LEN_MIN")
	minLength, err = strconv.Atoi(minLengthStr)
	if err != nil || minLength <= 0 {
		minLength = 10
	}

	maxLengthStr := os.Getenv("STRING_LEN_MAX")
	maxLength, err = strconv.Atoi(maxLengthStr)
	if err != nil || maxLength < minLength {
		maxLength = minLength + 10
	}

	// TODO: reconnect on connection loss. To reproduce: restart encryptor microservice and ws would not reconnect
	conn, _, _, err = ws.DefaultDialer.Dial(context.Background(), encEndpoint)
	if err != nil {
		fmt.Printf("Cannot connect to Encryptor microservice at %s.\n", encEndpoint)
		return
	}

	fmt.Printf("StringRandomizer microservice is listening on port %s.\n", svcPort)
	http.HandleFunc("/", randomizerHTTPServer)
	http.ListenAndServe(":"+svcPort, nil)
}
