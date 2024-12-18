package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

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

func setHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	util.IncreaseClock(&myClock, intNodeId)
	mutex.Unlock()
	// message := ...
	for node, _ := range allNodes {
		fmt.Println(node)
		// TODO message to node
	}
}

func main() {
	http.HandleFunc("/read", readHandler)
	http.HandleFunc("/set", setHandler)

	if err := http.ListenAndServe(nodeId, nil); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
