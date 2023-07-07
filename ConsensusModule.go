package raft

import (
	"fmt"
	"time"
)

func (c *ConsensusModule[j, x, k]) Vote(request RequestVote[j]) Reply {
	if c.VotedFor == -1 && request.Term >= c.CurrentTerm {
		nodeLastLogLen, nodeLastLogTerm := c.lastLog()
		if (request.LastLogIndex == nodeLastLogLen && nodeLastLogTerm == request.LastLogTerm) || request.LastLogIndex > nodeLastLogLen {
			c.VotedFor = int(request.CandidateId)
			return Reply{
				Term:        c.CurrentTerm,
				VoteGranted: true,
			}
		}
	}
	return Reply{
		Term:        c.CurrentTerm,
		VoteGranted: false,
	}
}

func (c *ConsensusModule[j, x, k]) AppendEntry(entries AppendEntries[j]) Reply {
	lastIndex, lastLog := c.lastLog()
	ll := c.Contact.GetLeaderLog()
	c.Contact.LogValue(ll)
	c.Contact.LogValue(c.Log)
	if entries.Term >= c.CurrentTerm && len(entries.Entries) == 0 && entries.PrevLogIndex == lastIndex && entries.PrevLogTerm == lastLog && c.Contact.ValidLog(ll) && c.Contact.LogValue(ll) == c.Contact.LogValue(c.Log) {
		c.CurrentTerm = entries.Term
		c.VotedFor = -1
		c.SetTicker()
		return Reply{
			Term:        c.CurrentTerm,
			VoteGranted: true,
		}
	} else if len(entries.Entries) > 0 && (entries.PrevLogIndex == lastIndex && entries.PrevLogTerm == lastLog) {
		for _, entry := range entries.Entries {
			if !c.Contact.ValidLogEntryCommand(entry.Command) {
				fmt.Println("Not valid command")
				return Reply{
					Term:        c.CurrentTerm,
					VoteGranted: false,
				}
			}
		}
		c.Log = append(c.Log, entries.Entries...)
		if c.State != Leader && ll != nil && len(ll) != len(c.Log) {
			// Need to add proper getting of missing logs
			c.Log = ll
		}
		fmt.Println(c.Id, "Log updated")
		return Reply{
			Term:        c.CurrentTerm,
			VoteGranted: true,
		}
	}
	return Reply{
		Term:        c.CurrentTerm,
		VoteGranted: false,
	}
}

func (c *ConsensusModule[j, x, k]) ResetTicker() {
	if c.Ticker == nil {
		c.Ticker = time.NewTicker(c.TickerDuration)
	} else {
		c.Ticker.Reset(c.TickerDuration)
	}
}
