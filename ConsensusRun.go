package raft

import "fmt"

func (c *ConsensusModule[j, k]) RunServer(done <-chan bool) {
main:
	for {
		select {
		case <-done:
			break main
		case <-*c.ReceiveChan:
			c.ResetTicker()
		case <-c.Ticker.C:
			if c.State == Follower {
				fmt.Println(c.Id, "ticker duration:", c.TickerDuration)
				c.State = Candidate
				c.CurrentTerm++
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
					} else {
					}
				}
				if vc > len(votes)/2 {
					c.State = Leader
					c.SetTicker()
					var heartbeat AppendEntries[j]
					if len(c.Log) == 0 {
						heartbeat = c.heartbeat(true)
					} else {
						heartbeat = c.heartbeat(false)
					}
					c.Contact.AppendEntries(heartbeat)
					fmt.Println(c.Id, "is leader ticks at", c.TickerDuration)
				}
			} else if c.State == Leader {
				var heartbeat AppendEntries[j]
				if len(c.Log) == 0 {
					heartbeat = c.heartbeat(true)
				} else {
					heartbeat = c.heartbeat(false)
				}
				c.Contact.AppendEntries(heartbeat)
			} else {
				c.State = Follower
			}
		}
	}
}
