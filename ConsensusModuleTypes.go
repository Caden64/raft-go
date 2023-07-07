package raft

import (
	"sync"
	"time"
)

type ConsensusModuleState int

const (
	Follower ConsensusModuleState = iota
	Candidate
	Leader
)

type Contact[j, x comparable, k any] interface {
	GetPeerIds() []uint
	GetLeader() uint
	GetLeaderLog() []LogEntry[j]
	RequestVotes(vote RequestVote[j]) []Reply
	AppendEntries(entries AppendEntries[j]) []Reply
	ValidLogEntryCommand(j) bool
	ValidLog([]LogEntry[j]) bool
	ExecuteLog(uint, []j) error
	DefaultLogEntryCommand() j
	LogValue([]LogEntry[j]) x
}

type LogEntry[j comparable] struct {
	Command j
	Term    uint
}

type RequestVote[j comparable] struct {
	Term         uint
	CandidateId  uint
	LastLogIndex int
	LastLogTerm  j
}

type Reply struct {
	Term        uint
	VoteGranted bool
}

type AppendEntries[j comparable] struct {
	Term         uint
	LeaderId     uint
	PrevLogIndex int
	PrevLogTerm  j
	Entries      []LogEntry[j]
	LeaderCommit uint
}

type ConsensusModule[j, x comparable, k any] struct {
	Mutex          *sync.Mutex
	Id             uint
	State          ConsensusModuleState
	Ticker         *time.Ticker
	TickerDuration time.Duration

	// Volatile state in memory
	LeaderId    uint
	CommitIndex uint
	LastApplied uint

	// Volatile state for leaders
	NextIndex  []uint
	MatchIndex []uint

	// Concurrent API communication
	ReceiveChan *chan k
	Contact     Contact[j, x, k]

	// Persistent state in memory
	CurrentTerm uint
	VotedFor    int
	Log         []LogEntry[j]
}
