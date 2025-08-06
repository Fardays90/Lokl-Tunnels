package main

import (
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var connections = make(map[string]*websocket.Conn)
var charset = "abcdefghijklmnoqprstyxz1234567890"
var serverPort string
var upgrader = websocket.Upgrader{}
var pendingResponses = sync.Map{}
var mutex = &sync.Mutex{}

type HTTPReq struct {
	ID      string              `json:"id"`
	Method  string              `json:"method"`
	Body    []byte              `json:"body"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers"`
}

type HTTPRes struct {
	ID         string              `json:"id"`
	StatusCode int                 `json:"status_code"`
	Body       []byte              `json:"body"`
	Headers    map[string][]string `json:"headers"`
}

type idJson struct {
	Id string `json:"id"`
}

var key string = loadEnv("tokens.env")

func generateRandomId(length int) string {
	id := make([]byte, length)
	for i := range id {
		id[i] = charset[mrand.Intn(len(charset))]
	}
	return string(id)
}
func loadEnv(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		log.Println("Can't open file err: " + err.Error())
		return ""
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatal(err.Error())
	}
	size := fileInfo.Size()
	requiredString := make([]byte, size)
	_, err = file.Read(requiredString)
	if err != nil {
		log.Fatal("Error while trying to read env file: " + err.Error())
	}
	secretKey := strings.Split(string(requiredString), "=")[1]
	return secretKey
}
func handleTunnelClient(w http.ResponseWriter, r *http.Request) {
	header := r.Header
	if header.Get("Authorization") != "Bearer "+key {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error trying to upgrade http connection to websocket err:" + err.Error())
		return
	}
	uniqueId := strings.ToLower(crand.Text())
	mutex.Lock()
	connections[uniqueId] = conn
	mutex.Unlock()
	uniqueIdJson := idJson{Id: uniqueId}
	conn.WriteJSON(uniqueIdJson)
	for {
		var response HTTPRes
		err := conn.ReadJSON(&response)
		if err != nil {
			fmt.Println("Error trying to read the json err: " + err.Error())
			mutex.Lock()
			delete(connections, uniqueId)
			mutex.Unlock()
			break
		}
		val, ok := pendingResponses.Load(response.ID)
		if !ok {
			fmt.Println("No pending response channel found for ID: ", response.ID)
			continue
		}
		responseChannel, ok := val.(chan HTTPRes)
		if !ok {
			fmt.Println("Pending response is not a channel for ID: ", response.ID)
			continue
		}
		responseChannel <- response
		pendingResponses.Delete(response.ID)
	}
}
func handleTest(w http.ResponseWriter, r *http.Request) {
	testMsg := idJson{Id: "Testing"}
	testJson, err := json.Marshal(testMsg)
	if err != nil {
		fmt.Println("Error trying to convert test to json")
		return
	}
	w.Write(testJson)
}
func handleExternalReqs(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	id := strings.Split(host, ".")[0]
	Path := r.URL.Path
	log.Println("Making req to -> " + id)
	mutex.Lock()
	tunnelToFind, found := connections[id]
	mutex.Unlock()
	if !found {
		log.Println("Tunnel does not exist or wrong id")
		return
	}
	method := r.Method
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body err:" + err.Error())
		return
	}
	defer r.Body.Close()
	requestHeaders := r.Header
	reqId := generateRandomId(6)
	clientReq := HTTPReq{ID: reqId, Method: method, Body: bodyBytes, Path: Path, Headers: requestHeaders}
	responseChan := make(chan HTTPRes)
	pendingResponses.Store(clientReq.ID, responseChan)
	tunnelToFind.WriteJSON(clientReq)
	response := <-responseChan
	for k, vv := range response.Headers {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(response.StatusCode)
	w.Write(response.Body)
}

func main() {
	http.HandleFunc("/connect", handleTunnelClient)
	http.HandleFunc("/", handleExternalReqs)
	http.HandleFunc("/test", handleTest)
	serverPort = ":8080"
	fmt.Println("Server started at localhost" + serverPort)
	err := http.ListenAndServe(serverPort, nil)
	if err != nil {
		fmt.Println("Error starting server at localhost" + serverPort)
		return
	}
}
