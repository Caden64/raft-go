package main

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

func main() {
	s1 := NewServer()
	election := s1.StartElection()
	s2 := NewServer()
	fmt.Println(s2.GiveElectionVote(election))
	s3 := NewServer()
	fmt.Println(s3.GiveElectionVote(election))
	/*
		main:
			for {
				select {
				case <-s1.Timeout.C:
					if s1.VotedFor == 0 {
						s1.StartElection()
						break main
					}
				}
			}
	*/
}

type Server struct {
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

func NewServer() Server {
	rn := rand.Intn(math.MaxInt)
	for rn == 0 {
		rn = rand.Intn(math.MaxInt)
	}
	return Server{
		Id:          rn,
		CurrentTerm: 0,
		VotedFor:    0,
		VoteTerm:    0,
		Log:         []int{},
		CommitIndex: 0,
		LastApplied: 0,
		NextIndex:   []int{},
		MatchIndex:  []int{},
		Timeout:     time.NewTicker(time.Millisecond * time.Duration(rand.Intn(999))),
		State:       Follower,
	}
}

func (s *Server) GiveElectionVote(request RequestVote) Response {
	var lastLogTerm bool

	if len(s.Log) > 0 && request.LastLogIndex > 1 {
		if s.Log[len(s.Log)-1] == request.LastLogTerm {
			lastLogTerm = true
		}
	} else if request.LastLogIndex == 1 {
		lastLogTerm = true
	}
	if request.Term <= s.CurrentTerm || (s.VotedFor != 0 && (s.VoteTerm <= request.Term || request.LastLogIndex != len(s.Log)+1 || !lastLogTerm)) {
		return Response{
			s.CurrentTerm,
			false,
		}
	}

	return Response{
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
