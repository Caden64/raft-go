package raft

func (c *ConsensusModule[j, k, x]) handleCandidate() {
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

func (c *ConsensusModule[j, k, x]) validCandidate() bool {
	var request RequestVote[j]
	if len(c.Log) == 0 {
		request = c.NewRequestVote(true)
	} else {
		request = c.NewRequestVote(false)
	}
	if c.VotedFor == -1 && request.Term >= c.CurrentTerm {
		nodeLastLogLen, nodeLastLogTerm := c.lastLog()
		if (request.LastLogIndex == nodeLastLogLen && nodeLastLogTerm == request.LastLogTerm) || request.LastLogIndex > nodeLastLogLen {
			return true
		}
	}
	return false
}
