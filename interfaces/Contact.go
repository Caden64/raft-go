package interfaces

import raft "raft-go"

type Contact[j any] interface {
	GetPeerIds() []uint
	RequestVotes(vote raft.RequestVote[j]) []raft.Reply
	AppendEntries(entries raft.AppendEntries[j]) []raft.Reply
}
