package raft

func (c *ConsensusModule[j, k]) followerToCandidate() {
	clear(c.MatchIndex)
	clear(c.NextIndex)
	if c.validCandidate() {
		c.State = Candidate
		c.CurrentTerm++
		c.handleCandidate()
	}
}
