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

var values map[string]Entry
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

	values = make(map[string]Entry)
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
		response := requests.Read{Value: value.value}
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
	client := &http.Client{
		Timeout: 500 * time.Millisecond,
	}
	resp, err := client.Post(*address, "application/json", bytes.NewBuffer(*payload))
	defer func() {
		if err == nil {
			resp.Body.Close()
		}
	}()
	if err != nil || resp.StatusCode == http.StatusForbidden {
		fmt.Println("Failed to send message to", *address, ", retrying in", retryTimeout)
		go func() {
			time.Sleep(retryTimeout)
			tryToSend(address, payload, retryTimeout*2)
		}()
		return
	}
	fmt.Println("Sent message to", *address)
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
		address := "http://" + node + "/broadcast"
		go tryToSend(&address, &payload, baseRetryTimeout)
	}

	w.WriteHeader(http.StatusOK)
}

func broadcastHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	if isStopped {
		mutex.Unlock()
		http.Error(w, "node is stopped", http.StatusForbidden)
		return
	}
	defer mutex.Unlock() // TODO сделать оптимально где надо

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var request requests.BroadcastRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	if len(request.Timestamp) != len(myClock.Value) {
		http.Error(w, "Timestamp length mismatch", http.StatusBadRequest)
		return
	}
	util.UpdateClock(myClock.Value, request.Timestamp)
	for key, value := range request.Values {
		currentValue, err := values[key]
		if err || len(currentValue.clock.Value) == 0 {
			values[key] = Entry{value: value, clock: util.VectorClock{Value: request.Timestamp}}
			continue
		}
		compareResult := util.CompareClock(currentValue.clock.Value, request.Timestamp)
		if compareResult == util.Less {
			values[key] = Entry{value: value, clock: util.VectorClock{Value: request.Timestamp}}
			continue
		}
		if compareResult == util.Parallel || compareResult == util.Equal {
			fmt.Println("!! Uncomparable clocks for key", key, ":")
			fmt.Println("\tmy value's:       ", currentValue.clock.Value)
			fmt.Println("\trequest's value's:", request.Timestamp)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/read", readHandler)
	http.HandleFunc("/set", setHandler)
	http.HandleFunc("/broadcast", broadcastHandler)

	if err := http.ListenAndServe(nodeId, nil); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
