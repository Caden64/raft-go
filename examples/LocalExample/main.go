package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	raft "raft-go"
)

func main() {
	cx := new(ContactExample[string, int, bool])
	s1 := raft.NewConsensusModule[string, int, bool](cx)
	s2 := raft.NewConsensusModule[string, int, bool](cx)
	s3 := raft.NewConsensusModule[string, int, bool](cx)
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
	entry := raft.AppendEntries[string]{
		Term:         elcx.CurrentTerm,
		LeaderId:     cx.Leader,
		PrevLogTerm:  elcx.Log[len(elcx.Log)-1].Command,
		PrevLogIndex: len(elcx.Log),
		Entries:      []raft.LogEntry[string]{{Term: elcx.CurrentTerm, Command: "SET 50"}},
	}
	elcx.AppendEntry(entry)
	for _, peer := range cx.Peers {
		peer.AppendEntry(entry)
	}
	wg.Wait()
}

type ContactExample[j string, x int, k bool] struct {
	Leader uint
	Peers  []*raft.ConsensusModule[j, x, k]
	Done   <-chan k
}

func (c *ContactExample[j, x, k]) AddPeer(module *raft.ConsensusModule[j, x, k]) {
	c.Peers = append(c.Peers, module)
}

func (c *ContactExample[j, x, k]) GetPeerIds() []uint {
	var final []uint
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, peer := range c.Peers {
		wg.Add(1)
		peer := peer
		go func(cm *raft.ConsensusModule[j, x, k]) {
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

func (c *ContactExample[j, x, k]) RequestVotes(vote raft.RequestVote[j]) []raft.Reply {
	var replies []raft.Reply
	for _, peer := range c.Peers {
		peer := peer
		voteResponse := peer.Vote(vote)
		replies = append(replies, voteResponse)
	}
	return replies
}

func (c *ContactExample[j, x, k]) AppendEntries(entries raft.AppendEntries[j]) []raft.Reply {
	var replies []raft.Reply
	cl := c.GetExactLeader().Log
	cl = append(cl, entries.Entries...)
	for _, peer := range c.Peers {
		if peer.State == raft.Leader {
			continue
		}
		appendResponse := peer.AppendEntry(entries)
		replies = append(replies, appendResponse)
	}
	return replies
}

func (c *ContactExample[j, x, k]) GetLeader() uint {
	for _, peer := range c.Peers {
		if peer.State == raft.Leader {
			return peer.Id
		}
	}
	return 0
}

func (c *ContactExample[j, x, k]) GetExactLeader() *raft.ConsensusModule[j, x, k] {
	for _, peer := range c.Peers {
		if peer.State == raft.Leader {
			return peer
		}
	}
	return nil
}

func (c *ContactExample[j, x, k]) ValidLogEntryCommand(operation j) bool {
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

func (c *ContactExample[j, x, k]) ExecuteLog(_ uint, _ []j) error {
	return nil
}

func (c *ContactExample[j, x, k]) DefaultLogEntryCommand() j {
	return "NEXT"
}

func (c *ContactExample[j, x, k]) GetLeaderLog() []raft.LogEntry[j] {
	return c.GetExactLeader().Log
}

func (c *ContactExample[j, x, k]) ValidLog(log []raft.LogEntry[j]) bool {
	final := true
	for _, item := range log {
		if !c.ValidLogEntryCommand(item.Command) {
			final = false
			break
		}
	}
	return final
}

func (c *ContactExample[j, x, k]) LogValue(log []raft.LogEntry[j]) x {
	total := 0
	for _, item := range log {
		if strings.HasPrefix(string(item.Command), "SET ") {
			values := strings.SplitN(string(item.Command), " ", 2)
			value, err := strconv.Atoi(values[1])
			if err != nil {
				return -1
			}
			total += value
		} else if string(item.Command) == "NEXT" {

		} else {
			return -1
		}
	}
	return x(total)
}
