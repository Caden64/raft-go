package raft

func (c *ConsensusModule[j, k, x]) RunServer(done <-chan bool) {
main:
	for {
		select {
		case <-done:
			break main
		case <-*c.ReceiveChan:
			c.ResetTicker()
		case <-c.Ticker.C:
			if c.State == Follower {
				c.followerToCandidate()
			} else if c.State == Leader {
				c.handleLeader()
			} else {
				c.State = Follower
			}
		}
	}
}
