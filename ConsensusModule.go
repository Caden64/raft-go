package raft

import "sync"

type ConsensusModuleState int

const (
	Follower ConsensusModuleState = iota
	Candidate
	Leader
)

func (s ConsensusModuleState) String() string {
	switch s {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		panic("Unknown State")
	}
}

type LogEntry[j any] struct {
	Command j
	Term    uint
}

type RequestVote struct {
	Term         uint
	CandidateId  uint
	LastLogIndex uint
	LastLogTerm  uint
}

type Reply struct {
	Term        uint
	VoteGranted bool
}

type AppendEntries[j any] struct {
	Term     uint
	LeaderId uint

	PrevLogIndex uint
	PrevLogTerm  uint
	Entries      []LogEntry[j]
	LeaderCommit uint
}
type Contact[j any] interface {
	GetPeerIds() []uint
	RequestVote(vote RequestVote) ([]Reply, error)
	AppendEntry(entries AppendEntries[j]) ([]Reply, error)
}

type ConsensusModule[j any] struct {
	mutex sync.Mutex
	id    uint
	state ConsensusModuleState

	Contact[j]

	Log []LogEntry[j]
}

func (receiver *ConsensusModule[j]) Get(index int) LogEntry[j] {
	return receiver.Log[index]
}

func (receiver *ConsensusModule[j]) Set(values []LogEntry[j]) {
	receiver.mutex.Lock()
	defer receiver.mutex.Unlock()
	receiver.Log = append(receiver.Log, values...)
}
