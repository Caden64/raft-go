package raft

func (c *ConsensusModule[j, k, x]) handleLeader() {
	heartbeat := c.heartbeat()
	c.Contact.AppendEntries(heartbeat)
}
