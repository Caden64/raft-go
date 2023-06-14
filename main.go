package main

import (
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

func main() {
	s := NewServer(true)
main:
	for {
		select {
		case <-s.Timeout.C:
			if s.VotedFor == 0 {
				s.StartElection()
				break main
			}
		}
	}
}

type Server struct {
	BootStrap   bool
	Id          int
	CurrentTerm int
	VotedFor    int
	VoteTerm    int
	Log         []LogEntry
	CommitIndex int
	LastApplied int
	NextIndex   []int
	MatchIndex  []int
	Timeout     *time.Ticker
	State       ServerState
}

type LogEntry struct {
	Term  int
	Index int
	Data  int
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
		Entries      LogEntry
		LeaderCommit int
	}
*/

type Response struct {
	Term        int
	VoteGranted bool
}

func NewServer(BootStrap bool) Server {
	rn := rand.Intn(math.MaxInt)
	for rn == 0 {
		rn = rand.Intn(math.MaxInt)
	}
	return Server{
		BootStrap:   BootStrap,
		Id:          rn,
		CurrentTerm: 0,
		VotedFor:    0,
		VoteTerm:    0,
		Log:         []LogEntry{},
		CommitIndex: 0,
		LastApplied: 0,
		NextIndex:   []int{},
		MatchIndex:  []int{},
		Timeout:     time.NewTicker(time.Millisecond * time.Duration(rand.Intn(999))),
		State:       Follower,
	}
}

func (s *Server) ReceiveVote(request RequestVote) Response {
	if request.Term <= s.CurrentTerm {
		return Response{
			s.CurrentTerm,
			false,
		}
	} else if s.VotedFor != 0 && s.VoteTerm <= request.Term {
		return Response{
			s.CurrentTerm,
			false,
		}
	}
	return Response{
		s.CurrentTerm,
		false,
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
			LastLogTerm:  s.Log[len(s.Log)-1].Term,
		}
	}
	return RequestVote{
		Term:         s.CurrentTerm,
		CandidateId:  s.Id,
		LastLogIndex: len(s.Log) + 1,
		LastLogTerm:  -1,
	}
}
