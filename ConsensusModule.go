package raft

import (
	"math/rand"
	"sync"
	"time"
)

type ConsensusModuleState int

const (
	Follower ConsensusModuleState = iota
	Candidate
	Leader
)

func (s ConsensusModuleState) String() string {
	switch s {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		panic("Unknown State")
	}
}

type LogEntry[j any] struct {
	Command j
	Term    uint
}

type RequestVote struct {
	Term         uint
	CandidateId  uint
	LastLogIndex uint
	LastLogTerm  uint
}

type Reply struct {
	Term        uint
	VoteGranted bool
}

type AppendEntries[j any] struct {
	Term     uint
	LeaderId uint

	PrevLogIndex uint
	PrevLogTerm  uint
	Entries      []LogEntry[j]
	LeaderCommit uint
}
type Contact[j any] interface {
	GetPeerIds() []uint
	RequestVotes(vote RequestVote) ([]Reply, error)
	AppendEntries(entries AppendEntries[j]) ([]Reply, error)
}

type ConsensusModule[j any] struct {
	Mutex sync.Mutex
	Id    uint
	// Volatile state in memory
	State          ConsensusModuleState
	Ticker         *time.Ticker
	TickerDuration time.Duration

	Contact[j]

	// Persistent state in memory
	CurrentTerm uint
	VotedFor    int
	Log         []LogEntry[j]
}

func (c *ConsensusModule[j]) ResetTicker() {
	c.Ticker = time.NewTicker(c.TickerDuration)
}
func (c *ConsensusModule[j]) Get(index int) LogEntry[j] {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	return c.Log[index]
}

func (c *ConsensusModule[j]) Set(values []LogEntry[j]) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.Log = append(c.Log, values...)
}

func (c *ConsensusModule[j]) SetTicker() {
	if c.State != Leader {
		ri := rand.Intn(500)
		for ri < 250 {
			ri = rand.Intn(500)
		}
		c.TickerDuration = time.Duration(ri) * time.Millisecond
	} else {
		ri := rand.Intn(150)
		for ri < 50 {
			ri = rand.Intn(150)
		}
		c.TickerDuration = time.Duration(ri) * time.Millisecond
	}

	c.ResetTicker()
}
