package raft

import (
	"math/rand"
	"sync"
	"time"
)

func (c *ConsensusModule[j, x, k]) NewHeartbeat() AppendEntries[j] {
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

func (c *ConsensusModule[j, x, k]) lastLog() (int, j) {
	if len(c.Log) == 0 {
		return 1, *new(j)
	} else {
		return len(c.Log), c.Log[len(c.Log)-1].Command
	}
}
