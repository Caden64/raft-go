package server

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type ServerState int

const (
	Follower ServerState = iota
	Candidate
	Leader
)

type Server struct {
	Name        string
	Id          int
	CurrentTerm int
	VotedFor    int
	VoteTerm    int
	Log         []int
	CommitIndex int
	LastApplied int
	NextIndex   []int
	MatchIndex  []int
	Timeout     *time.Ticker
	State       ServerState
	ServerCount int
}

type RequestVote struct {
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

/*
	type AppendLog struct {
		Term         int
		LeaderId     int
		PrevLogIndex int
		PrevLogTerm  int
		Entries      []int
		LeaderCommit int
	}
*/

type VoteResponse struct {
	Term        int
	VoteGranted bool
}

func NewServer(name string) Server {
	rn := rand.Intn(math.MaxInt)
	for rn == 0 {
		rn = rand.Intn(math.MaxInt)
	}
	return Server{
		Name:        name,
		Id:          rn,
		CurrentTerm: 0,
		VotedFor:    0,
		VoteTerm:    0,
		Log:         []int{},
		CommitIndex: 1,
		LastApplied: 1,
		NextIndex:   []int{},
		MatchIndex:  []int{},
		Timeout:     time.NewTicker(time.Millisecond * time.Duration(rand.Intn(999))),
		State:       Follower,
	}
}

func (s *Server) GiveElectionVote(request RequestVote) VoteResponse {
	var lastLogTerm bool

	if len(s.Log) > 0 && request.LastLogIndex > 1 {
		if s.Log[len(s.Log)-1] == request.LastLogTerm {
			lastLogTerm = true
		}
	} else if request.LastLogIndex == 1 {
		lastLogTerm = true
	}
	if request.Term <= s.CurrentTerm || (s.VotedFor != 0 && (s.VoteTerm <= request.Term || request.LastLogIndex != len(s.Log)+1 || !lastLogTerm)) {
		return VoteResponse{
			s.CurrentTerm,
			false,
		}
	}

	return VoteResponse{
		request.Term,
		true,
	}
}

func (s *Server) StartElection() RequestVote {
	s.State = Candidate
	s.CurrentTerm++
	if len(s.Log) > 0 {
		return RequestVote{
			Term:         s.CurrentTerm,
			CandidateId:  s.Id,
			LastLogIndex: len(s.Log) + 1,
			LastLogTerm:  s.Log[len(s.Log)-1],
		}
	}
	return RequestVote{
		Term:         s.CurrentTerm,
		CandidateId:  s.Id,
		LastLogIndex: len(s.Log) + 1,
		LastLogTerm:  -1,
	}
}

func (s *Server) SendElection(sc serverContact, rv RequestVote) []VoteResponse {
	return sc.SendRequestVote(rv)
}

func (s *Server) PromoteLeader(sc ServerCount, responses ...VoteResponse) bool {
	total := sc.TotalServerCount()
	totalVotes := 0
	for _, response := range responses {
		if response.VoteGranted {
			totalVotes++
		}
	}

	if totalVotes > total/2 {
		s.State = Leader
		fmt.Println("Leader")
		return true
	}
	return false
}

type ServerCount interface {
	TotalServerCount() int
}

type serverContact interface {
	SendRequestVote(vote RequestVote) []VoteResponse
}
