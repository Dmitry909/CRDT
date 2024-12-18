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

func CompareClock(lhs []int, rhs []int) CompareResult {
	if len(lhs) != len(rhs) {
		log.Fatal("CompareClock called with clocks of different lengths")
	}
	if len(lhs) == 0 {
		log.Fatal("CompareClock called with clock of 0 lengths")
	}

	couldBeLess := true
	couldBeEqual := true
	couldBeMore := true
	for _, i := range lhs {
		if lhs[i] < rhs[i] {
			couldBeEqual = false
			couldBeMore = false
		} else if lhs[i] > rhs[i] {
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

func UpdateClock(to *VectorClock, from []int) {
	for _, i := range to.Value {
		if to.Value[i] < from[i] {
			to.Value[i] = from[i]
		}
	}
}
