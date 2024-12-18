package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

	"CRDT/requests"
	"CRDT/util"
)

type Clock struct {
	// ??? TODO
}

type Entry struct {
	value string
	clock Clock
}

var values map[string]string
var isStopped bool
var mutex sync.Mutex
var port string
var nodeId string
var allNodes = []string{}
var nodesExceptMe = []string{}

func init() {
	port = os.Args[1]
	nodeId = util.LocalIP + ":" + port

	allNodes = util.ConvertPortsToSlice(os.Args[2])

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
	// TODO
}

func main() {
	http.HandleFunc("/read", readHandler)
	http.HandleFunc("/set", setHandler)

	if err := http.ListenAndServe(nodeId, nil); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
