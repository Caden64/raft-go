package raft

func (c *ConsensusModule[j, k]) handleLeader() {
	heartbeat := c.heartbeat()
	c.Contact.AppendEntries(heartbeat)
}
