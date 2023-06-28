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

type Contact[j any, k any] interface {
	GetPeerIds() []uint
	RequestVotes(vote RequestVote[j]) []Reply
	AppendEntries(entries AppendEntries[j]) []Reply
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
		panic("Unknown State")
	}
}

type LogEntry[j any] struct {
	Command j
	Term    uint
}

type RequestVote[j any] struct {
	Term         uint
	CandidateId  uint
	LastLogIndex uint
	LastLogTerm  j
}

type Reply struct {
	Term        uint
	VoteGranted bool
}

type AppendEntries[j any] struct {
	Term     uint
	LeaderId uint

	PrevLogIndex uint
	PrevLogTerm  j
	Entries      []LogEntry[j]
}

type ConsensusModule[j any, k any] struct {
	Mutex *sync.Mutex
	Id    uint
	// Volatile state in memory
	State          ConsensusModuleState
	Ticker         *time.Ticker
	TickerDuration time.Duration

	ReceiveChan *chan k

	Contact Contact[j, k]

	// Persistent state in memory
	CurrentTerm uint
	VotedFor    int
	Log         []LogEntry[j]
}

func NewConsensusModule[j any, k any](contact Contact[j, k]) *ConsensusModule[j, k] {
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
	c.Ticker = time.NewTicker(c.TickerDuration)
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

func (c *ConsensusModule[j, k]) Vote(request RequestVote[j]) Reply {
	fmt.Println("Gave vote to: ", request.CandidateId, "From: ", c.Id)
	return Reply{
		Term:        c.CurrentTerm,
		VoteGranted: true,
	}
}

func (c *ConsensusModule[j, k]) AppendEntry(entry AppendEntries[j]) Reply {
	if len(entry.Entries) == 0 {
		c.ResetTicker()
	}
	return Reply{
		Term:        c.CurrentTerm,
		VoteGranted: true,
	}
}
func (c *ConsensusModule[j, k]) RunServer(done <-chan bool) {
main:
	for {
		select {
		case <-done:
			break main
		case <-*c.ReceiveChan:
			c.ResetTicker()
		case <-c.Ticker.C:
			if c.State == Follower {
				c.State = Candidate
				c.CurrentTerm++
				fmt.Println("Vote: ", c.Id)
				var serverRequestVote RequestVote[j]
				if len(c.Log) == 0 {
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
						LastLogIndex: uint(len(c.Log) + 1),
						LastLogTerm:  c.Log[len(c.Log)-1].Command,
					}
				}
				votes := c.Contact.RequestVotes(serverRequestVote)
				fmt.Println("Got votes", c.Id)
				var vc int
				for _, vote := range votes {
					if vote.VoteGranted {
						vc++
					}
				}
				if vc > len(votes)/2 {
					c.State = Leader
					c.SetTicker()
					fmt.Println("Should be leader")
					var heartbeat AppendEntries[j]
					if len(c.Log) == 0 {
						heartbeat = AppendEntries[j]{
							Term:         c.CurrentTerm,
							LeaderId:     c.Id,
							PrevLogIndex: 1,
							PrevLogTerm:  *new(j),
							Entries:      []LogEntry[j]{},
						}
					} else {
						heartbeat = AppendEntries[j]{
							Term:         c.CurrentTerm,
							LeaderId:     c.Id,
							PrevLogIndex: uint(len(c.Log) + 1),
							PrevLogTerm:  c.Log[len(c.Log)-1].Command,
							Entries:      []LogEntry[j]{},
						}
					}
					c.Contact.AppendEntries(heartbeat)
					fmt.Println("Should be leader always")
				}
			} else if c.State == Leader {
				var heartbeat AppendEntries[j]
				if len(c.Log) == 0 {
					heartbeat = AppendEntries[j]{
						Term:         c.CurrentTerm,
						LeaderId:     c.Id,
						PrevLogIndex: 1,
						PrevLogTerm:  *new(j),
						Entries:      []LogEntry[j]{},
					}
				} else {
					heartbeat = AppendEntries[j]{
						Term:         c.CurrentTerm,
						LeaderId:     c.Id,
						PrevLogIndex: uint(len(c.Log) + 1),
						PrevLogTerm:  c.Log[len(c.Log)-1].Command,
						Entries:      []LogEntry[j]{},
					}
				}
				c.Contact.AppendEntries(heartbeat)
			}
		}
	}
}
