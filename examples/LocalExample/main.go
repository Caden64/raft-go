package main

import (
	"fmt"
	"sync"

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
	for _, peer := range cx.GetPeerIds() {
		fmt.Println(peer)
	}
	var wg sync.WaitGroup
	wg.Add(3)
	go s1.RunServer(cx.Done)
	go s2.RunServer(cx.Done)
	go s3.RunServer(cx.Done)
	wg.Wait()
}

type ContactExample[j comparable, k any] struct {
	Peers []*raft.ConsensusModule[j, k]
	Done  <-chan k
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
	fmt.Println("Giving votes to node", vote.CandidateId)
	return replies
}

func (c *ContactExample[j, k]) AppendEntries(entries raft.AppendEntries[j]) []raft.Reply {
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
