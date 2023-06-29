package raft

func (c *ConsensusModule[j, k]) followerToCandidate() {
	c.State = Candidate
	c.CurrentTerm++
	c.handleCandidate()
}
