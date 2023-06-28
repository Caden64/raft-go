package main

import (
	"fmt"
	raft "raft-go"
	"sync"
)

func main() {
	cl := new(ContactLol[string, bool])
	s1 := raft.NewConsensusModule[string, bool](cl)
	s2 := raft.NewConsensusModule[string, bool](cl)
	s3 := raft.NewConsensusModule[string, bool](cl)
	cl.AddPeer(s1)
	cl.AddPeer(s2)
	cl.AddPeer(s3)
	for _, peer := range cl.GetPeerIds() {
		fmt.Println(peer)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go s1.RunServer(cl.Done)
	go s2.RunServer(cl.Done)
	go s3.RunServer(cl.Done)
	wg.Wait()
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
	for _, peer := range c.Peers {
		if peer.Id == vote.CandidateId {
			continue
		}
		peer := peer
		voteResponse := peer.Vote(vote)
		replies = append(replies, voteResponse)
	}
	fmt.Println("Giving votes to node")
	return replies
}

func (c *ContactLol[j, k]) AppendEntries(entries raft.AppendEntries[j]) []raft.Reply {
	var replies []raft.Reply
	for _, peer := range c.Peers {
		if peer.Id == entries.LeaderId {
			continue
		}
		peer := peer
		appendResponse := peer.AppendEntry(entries)
		replies = append(replies, appendResponse)
	}
	return replies
}
