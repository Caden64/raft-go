package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	raft "raft-go"
)

func main() {
	cx := new(ContactExample[string, bool])
	s1 := raft.NewConsensusModule[string, bool](cx)
	s2 := raft.NewConsensusModule[string, bool](cx)
	s3 := raft.NewConsensusModule[string, bool](cx)
	cx.AddPeer(s1)
	cx.AddPeer(s2)
	cx.AddPeer(s3)
	var wg sync.WaitGroup
	wg.Add(3)
	go s1.RunServer(cx.Done)
	go s2.RunServer(cx.Done)
	go s3.RunServer(cx.Done)

	time.Sleep(time.Second * 1)
	cx.Leader = cx.GetLeader()
	fmt.Println(cx.Leader)
	elcx := cx.GetExactLeader()
	for _, peer := range cx.Peers {
		peer.AppendEntry(raft.AppendEntries[string]{
			Term:         elcx.CurrentTerm,
			LeaderId:     cx.Leader,
			PrevLogTerm:  "",
			PrevLogIndex: 1,
			Entries:      []raft.LogEntry[string]{{Term: elcx.CurrentTerm, Command: "SET 50"}},
		},
		)
	}
	wg.Wait()
}

type ContactExample[j string, k bool] struct {
	Leader uint
	Peers  []*raft.ConsensusModule[j, k]
	Done   <-chan k
}

func (c *ContactExample[j, k]) AddPeer(module *raft.ConsensusModule[j, k]) {
	c.Peers = append(c.Peers, module)
}

func (c *ContactExample[j, k]) GetPeerIds() []uint {
	var final []uint
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, peer := range c.Peers {
		wg.Add(1)
		peer := peer
		go func(cm *raft.ConsensusModule[j, k]) {
			cm.Mutex.Lock()
			mu.Lock()
			defer mu.Unlock()
			defer wg.Done()
			defer cm.Mutex.Unlock()
			final = append(final, peer.Id)
		}(peer)
	}
	wg.Wait()
	return final
}

func (c *ContactExample[j, k]) RequestVotes(vote raft.RequestVote[j]) []raft.Reply {
	var replies []raft.Reply
	for _, peer := range c.Peers {
		peer := peer
		voteResponse := peer.Vote(vote)
		replies = append(replies, voteResponse)
	}
	return replies
}

func (c *ContactExample[j, k]) AppendEntries(entries raft.AppendEntries[j]) []raft.Reply {
	var replies []raft.Reply
	// cl := c.GetExactLeader().Log
	// fmt.Println(checkLog)
	for _, peer := range c.Peers {
		peer := peer
		appendResponse := peer.AppendEntry(entries)
		replies = append(replies, appendResponse)
	}
	return replies
}

func (c *ContactExample[j, k]) GetLeader() uint {
	for _, peer := range c.Peers {
		if peer.State == raft.Leader {
			return peer.Id
		}
	}
	return 0
}

func (c *ContactExample[j, k]) GetExactLeader() *raft.ConsensusModule[j, k] {
	for _, peer := range c.Peers {
		if peer.State == raft.Leader {
			return peer
		}
	}
	return nil
}

func (c *ContactExample[j, k]) ValidLogEntryCommand(operation j) bool {
	if strings.HasPrefix(string(operation), "SET") && len(operation) >= 5 {
		operationValues := strings.SplitN(string(operation), " ", 2)
		if len(operationValues) > 1 {
			_, err := strconv.Atoi(operationValues[1])
			if err == nil {
				return true
			}
		}
	} else if string(operation) == "NEXT" {
		return true
	}
	return false
}

func (c *ContactExample[j, k]) ExecuteLogEntryCommand(server uint, operation j) error {
	if !c.ValidLogEntryCommand(operation) {
		return errors.New("invalid log entry command")
	}
	for _, peer := range c.Peers {
		peer := peer
		if peer.Id == server {
			return func() error {
				peer.Mutex.Lock()
				defer peer.Mutex.Unlock()
				logEntry := raft.LogEntry[j]{
					Command: operation,
					Term:    peer.CurrentTerm,
				}
				peer.Log = append(peer.Log, logEntry)
				if strings.HasPrefix(string(operation), "SET") {
					logEntry.Command = "NEXT"
					peer.Log = append(peer.Log, logEntry)
				}
				return nil
			}()
		}
	}
	return errors.New("unable to find node")
}

func (c *ContactExample[j, k]) DefaultLogEntryCommand() j {
	return "NEXT"
}

func (c *ContactExample[j, k]) GetLeaderLog() []raft.LogEntry[j] {
	leader := c.GetLeader()
	for _, peer := range c.Peers {
		peer := peer
		if peer.Id == leader {
			return peer.Log
		}
	}
	return nil
}
