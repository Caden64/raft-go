package main

import (
	raft "raft-go"
	"sync"
)

func main() {
	cl := new(ContactLol[string, bool])
	s1 := raft.NewConsensusModule[string, bool](cl)
	s1.RunServer(cl.Done)
}

type ContactLol[j any, k any] struct {
	Peers []*raft.ConsensusModule[j, k]
	Done  <-chan k
}

func (c *ContactLol[j, k]) AddPeer(module *raft.ConsensusModule[j, k]) {
	c.Peers = append(c.Peers, module)
}

func (c *ContactLol[j, k]) GetPeerIds() []uint {
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

func (c *ContactLol[j, k]) RequestVotes(vote raft.RequestVote[j]) []raft.Reply {
	var replies []raft.Reply
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
			defer mu.Lock()
			voteResponse := cm.Vote(vote)
			replies = append(replies, voteResponse)
		}(peer)
	}
	return replies
}

func (c *ContactLol[j, k]) AppendEntries(entries raft.AppendEntries[j]) []raft.Reply {
	var replies []raft.Reply
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
			defer mu.Lock()
			appendResponse := cm.AppendEntry(entries)
			replies = append(replies, appendResponse)
		}(peer)
	}
	return replies
}
