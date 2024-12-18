package util

import (
	"log"
	"strings"
)

var LocalIP = "127.0.0.1"

func ConvertPortsToSlice(ports string) []string {
	portList := strings.Split(ports, ",")
	hostPorts := make([]string, len(portList))

	for i, port := range portList {
		hostPorts[i] = LocalIP + ":" + strings.TrimSpace(port)
	}

	return hostPorts
}

type VectorClock struct {
	Value []int
}

type CompareResult int

const (
	Less CompareResult = iota
	Equal
	More
	Parallel
)

func CompareClock(lhs *VectorClock, rhs *VectorClock) CompareResult {
	if len(lhs.Value) != len(rhs.Value) {
		log.Fatal("CompareClock called with clocks of different lengths")
	}
	if len(lhs.Value) == 0 {
		log.Fatal("CompareClock called with clock of 0 lengths")
	}

	couldBeLess := true
	couldBeEqual := true
	couldBeMore := true
	for _, i := range lhs.Value {
		if lhs.Value[i] < rhs.Value[i] {
			couldBeEqual = false
			couldBeMore = false
		} else if lhs.Value[i] > rhs.Value[i] {
			couldBeEqual = false
			couldBeLess = false
		}
	}

	if couldBeEqual {
		return Equal
	}
	if couldBeLess {
		return Less
	}
	if couldBeMore {
		return More
	}
	return Parallel
}

func IncreaseClock(clock *VectorClock, intNodeId int) {
	if intNodeId < 0 || intNodeId >= len(clock.Value) {
		log.Fatal("IncreaseClock called with wrong intNodeId")
	}
	clock.Value[intNodeId]++
}
