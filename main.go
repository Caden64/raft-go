package main

import (
	"fmt"
	"math/rand"
	"time"
)

type ServerState int

const (
	Follower ServerState = iota
	Candidate
	Leader
)

func main() {
	t := time.NewTicker(time.Millisecond * time.Duration(rand.Intn(999)))
	<-t.C
	fmt.Println("ticked")
}

type Server struct {
	Id          int
	CurrentTerm int
	VotedFor    int
	Log         []int
	CommitIndex int
	LastApplied int
	NextIndex   []int
	MatchIndex  []int
	State       ServerState
}

type LogEntry struct {
	Term  int
	Index int
	Data  int
}
