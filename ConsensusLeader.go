package raft

func (c *ConsensusModule[j, k, x]) handleLeader() {
	heartbeat := c.NewHeartbeat()
	c.Contact.AppendEntries(heartbeat)
}
