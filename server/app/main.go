package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"CRDT/requests"
	"CRDT/util"
)

type Entry struct {
	value string
	clock util.VectorClock
}

var mutex sync.Mutex
var intNodeId int
var port string
var nodeId string
var allNodes = []string{}
var nodesExceptMe = []string{}

var values map[string]string
var isStopped bool
var myClock util.VectorClock

var baseRetryTimeout = 500 * time.Millisecond

func init() {
	intNodeId, _ = strconv.Atoi(os.Args[1])

	port = os.Args[2]
	nodeId = util.LocalIP + ":" + port

	allNodes = util.ConvertPortsToSlice(os.Args[3])

	contains := false
	for _, address := range allNodes {
		if address == nodeId {
			contains = true
		} else {
			nodesExceptMe = append(nodesExceptMe, address)
		}
	}
	if !contains {
		log.Fatal("Wrong port " + port)
	}

	values = make(map[string]string)
	isStopped = false
	myClock = util.VectorClock{Value: make([]int, len(allNodes))}
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	if isStopped {
		mutex.Unlock()
		http.Error(w, "node is stopped", http.StatusForbidden)
		return
	}
	mutex.Unlock()

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	value, ok := values[key]
	mutex.Unlock()
	if ok {
		response := requests.Read{Value: value}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	http.Error(w, "key not found", http.StatusNotFound)
}

func tryToSend(address *string, payload *[]byte, retryTimeout time.Duration) {
	resp, err := http.Post(*address, "application/json", bytes.NewBuffer(*payload))
	defer func() {
		if err == nil {
			resp.Body.Close()
		}
	}()
	if err != nil || resp.StatusCode == http.StatusForbidden {
		go func() {
			time.Sleep(retryTimeout)
			tryToSend(address, payload, retryTimeout*2)
		}()
	}
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	// TODO check PATCH request.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var request requests.SetRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	util.IncreaseClock(&myClock, intNodeId)
	broadcastRequest := map[string]interface{}{
		"values":    request.Values,
		"timestamp": myClock.Value,
	}
	payload, err := json.Marshal(broadcastRequest)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}
	mutex.Unlock()

	for _, node := range allNodes {
		address := node + "/broadcast"
		go tryToSend(&address, &payload, baseRetryTimeout)
	}

	w.WriteHeader(http.StatusOK)
}

func broadcastHandler(w http.ResponseWriter, r *http.Request) {
	// TODO implement
}

func main() {
	http.HandleFunc("/read", readHandler)
	http.HandleFunc("/set", setHandler)
	http.HandleFunc("/broadcast", broadcastHandler)

	if err := http.ListenAndServe(nodeId, nil); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
