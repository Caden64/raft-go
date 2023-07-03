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

type Contact[j comparable, k any] interface {
	GetPeerIds() []uint
	GetLeader() uint
	RequestVotes(vote RequestVote[j]) []Reply
	AppendEntries(entries AppendEntries[j]) []Reply
	ValidLogEntryCommand(j) bool
	ExecuteLogEntryCommand(uint, j) error
	DefaultLogEntryCommand() j
}

func (s ConsensusModuleState) String() string {
	switch s {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		return "Invalid"
	}
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

type ConsensusModule[j comparable, k any] struct {
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
	Contact     Contact[j, k]

	// Persistent state in memory
	CurrentTerm uint
	VotedFor    int
	Log         []LogEntry[j]
}

func NewConsensusModule[j comparable, k any](contact Contact[j, k]) *ConsensusModule[j, k] {
	cm := &ConsensusModule[j, k]{
		Mutex: new(sync.Mutex),
		Id:    uint(rand.Uint64()),
		State: Follower,

		ReceiveChan: new(chan k),
		Contact:     contact,

		CurrentTerm: 0,
		VotedFor:    -1,
		Log:         *new([]LogEntry[j]),
	}
	cm.SetTicker()
	cm.ResetTicker()
	return cm
}

func (c *ConsensusModule[j, k]) ResetTicker() {
	// fmt.Println(c.Id, " Ticker reset")
	if c.Ticker == nil {
		c.Ticker = time.NewTicker(c.TickerDuration)
	} else {
		c.Ticker.Reset(c.TickerDuration)
	}
}

func (c *ConsensusModule[j, k]) Get(index int) LogEntry[j] {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	return c.Log[index]
}

func (c *ConsensusModule[j, k]) Set(values []LogEntry[j]) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.Log = append(c.Log, values...)
}

func (c *ConsensusModule[j, k]) SetTicker() {
	if c.State != Leader {
		ri := rand.Intn(450)
		for ri < 250 {
			ri = rand.Intn(450)
		}
		c.TickerDuration = time.Duration(ri) * time.Millisecond
	} else {
		ri := rand.Intn(200)
		for ri < 50 {
			ri = rand.Intn(200)
		}
		c.TickerDuration = time.Duration(ri) * time.Millisecond
	}

	c.ResetTicker()
}

func (c *ConsensusModule[j, k]) Vote(request RequestVote[j]) Reply {
	if c.VotedFor == -1 && request.Term >= c.CurrentTerm {
		nodeLastLogLen, nodeLastLogTerm := c.lastLog()
		if (request.LastLogIndex == nodeLastLogLen && nodeLastLogTerm == request.LastLogTerm) || request.LastLogIndex > nodeLastLogLen {
			c.VotedFor = int(request.CandidateId)
			return Reply{
				Term:        c.CurrentTerm,
				VoteGranted: true,
			}
		}
	}
	return Reply{
		Term:        c.CurrentTerm,
		VoteGranted: false,
	}
}

func (c *ConsensusModule[j, k]) lastLog() (int, j) {
	if len(c.Log) == 0 {
		return 0, *new(j)
	} else {
		return len(c.Log), c.Log[len(c.Log)-1].Command
	}
}

func (c *ConsensusModule[j, k]) AppendEntry(entry AppendEntries[j]) Reply {
	if entry.Term >= c.CurrentTerm && len(entry.Entries) == 0 && entry.PrevLogIndex == -1 && entry.PrevLogTerm == *new(j) {
		c.CurrentTerm = entry.Term
		c.VotedFor = -1
		c.SetTicker()
		return Reply{
			Term:        c.CurrentTerm,
			VoteGranted: true,
		}
	}
	return Reply{
		Term:        c.CurrentTerm,
		VoteGranted: false,
	}
}

func (c *ConsensusModule[j, k]) heartbeat() AppendEntries[j] {
	return AppendEntries[j]{
		Term:         c.CurrentTerm,
		LeaderId:     c.Id,
		PrevLogIndex: -1,
		PrevLogTerm:  *new(j),
		Entries:      []LogEntry[j]{},
		LeaderCommit: uint(len(c.Log)),
	}
}

func (c *ConsensusModule[j, k]) NewRequestVote(newCM bool) RequestVote[j] {
	var serverRequestVote RequestVote[j]
	if newCM {
		serverRequestVote = RequestVote[j]{
			Term:         c.CurrentTerm,
			CandidateId:  c.Id,
			LastLogIndex: 1,
			LastLogTerm:  *new(j),
		}
	} else {
		serverRequestVote = RequestVote[j]{
			Term:         c.CurrentTerm,
			CandidateId:  c.Id,
			LastLogIndex: len(c.Log) + 1,
			LastLogTerm:  c.Log[len(c.Log)-1].Command,
		}
	}

	return serverRequestVote
}
