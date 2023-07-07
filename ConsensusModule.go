package raft

import (
	"fmt"
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

func NewConsensusModule[j, x comparable, k any](contact Contact[j, x, k]) *ConsensusModule[j, x, k] {
	cm := &ConsensusModule[j, x, k]{
		Mutex: new(sync.Mutex),
		Id:    uint(rand.Uint64()),
		State: Follower,

		ReceiveChan: new(chan k),
		Contact:     contact,

		CurrentTerm: 0,
		VotedFor:    -1,
		Log: []LogEntry[j]{
			{
				Command: contact.DefaultLogEntryCommand(),
				Term:    0,
			},
		},
	}
	cm.SetTicker()
	cm.ResetTicker()
	return cm
}

func (c *ConsensusModule[j, x, k]) ResetTicker() {
	if c.Ticker == nil {
		c.Ticker = time.NewTicker(c.TickerDuration)
	} else {
		c.Ticker.Reset(c.TickerDuration)
	}
}

func (c *ConsensusModule[j, x, k]) SetTicker() {
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

func (c *ConsensusModule[j, x, k]) Vote(request RequestVote[j]) Reply {
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

func (c *ConsensusModule[j, x, k]) lastLog() (int, j) {
	if len(c.Log) == 0 {
		return 1, *new(j)
	} else {
		return len(c.Log), c.Log[len(c.Log)-1].Command
	}
}

func (c *ConsensusModule[j, x, k]) AppendEntry(entries AppendEntries[j], ll []LogEntry[j]) Reply {
	lastIndex, lastLog := c.lastLog()

	c.Contact.LogValue(ll)
	c.Contact.LogValue(c.Log)
	if entries.Term >= c.CurrentTerm && len(entries.Entries) == 0 && entries.PrevLogIndex == lastIndex && entries.PrevLogTerm == lastLog && c.Contact.ValidLog(ll) && c.Contact.LogValue(ll) == c.Contact.LogValue(c.Log) {
		c.CurrentTerm = entries.Term
		c.VotedFor = -1
		c.SetTicker()
		return Reply{
			Term:        c.CurrentTerm,
			VoteGranted: true,
		}
	} else if len(entries.Entries) > 0 && (entries.PrevLogIndex == lastIndex && entries.PrevLogTerm == lastLog) {
		for _, entry := range entries.Entries {
			if !c.Contact.ValidLogEntryCommand(entry.Command) {
				fmt.Println("Not valid command")
				return Reply{
					Term:        c.CurrentTerm,
					VoteGranted: false,
				}
			}
		}
		c.Log = append(c.Log, entries.Entries...)
		if c.State != Leader && ll != nil && len(ll) != len(c.Log) {
			c.Log = ll
		}
		fmt.Println(c.Id, "Log updated")
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

func (c *ConsensusModule[j, x, k]) heartbeat() AppendEntries[j] {
	return AppendEntries[j]{
		Term:         c.CurrentTerm,
		LeaderId:     c.Id,
		PrevLogIndex: -1,
		PrevLogTerm:  *new(j),
		Entries:      []LogEntry[j]{},
		LeaderCommit: uint(len(c.Log)),
	}
}

func (c *ConsensusModule[j, x, k]) NewRequestVote(newCM bool) RequestVote[j] {
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
