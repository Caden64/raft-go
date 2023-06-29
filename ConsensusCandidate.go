package raft

import "fmt"

func (c *ConsensusModule[j, k]) handleCandidate() {
	fmt.Println(c.Id, "started vote")
	var serverRequestVote RequestVote[j]
	if len(c.Log) == 0 {
		serverRequestVote = c.NewRequestVote(true)
	} else {
		serverRequestVote = c.NewRequestVote(false)
	}
	votes := c.Contact.RequestVotes(serverRequestVote)
	var vc int
	for _, vote := range votes {
		if vote.VoteGranted {
			vc++
		}
	}
	if vc > len(votes)/2 {
		c.State = Leader
		c.SetTicker()
	}
}
