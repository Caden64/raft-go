package server

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type State int

const (
	Follower State = iota
	Candidate
	Leader
)

type Server struct {
	Name            string
	Id              int
	CurrentTerm     int
	VotedFor        int
	VoteTerm        int
	Log             []int
	CommitIndex     int
	LastApplied     int
	NextIndex       []int
	MatchIndex      []int
	Timeout         *time.Ticker
	TimeOutDuration time.Duration
	State           State
	ServerCount     int
}

type RequestVote struct {
	Name         string
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

type AppendLog struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []int
	LeaderCommit int
}

type Response struct {
	Term        int
	VoteGranted bool
}

func NewServer(name string) Server {
	rn := rand.Intn(math.MaxInt)
	for rn == 0 {
		rn = rand.Intn(math.MaxInt)
	}
	dur := time.Millisecond * time.Duration(rand.Intn(999))
	for dur < 50*time.Millisecond {
		dur = time.Millisecond * time.Duration(rand.Intn(999))
	}
	return Server{
		Name:            name,
		Id:              rn,
		CurrentTerm:     0,
		VotedFor:        0,
		VoteTerm:        0,
		Log:             []int{},
		CommitIndex:     1,
		LastApplied:     1,
		NextIndex:       []int{},
		MatchIndex:      []int{},
		Timeout:         time.NewTicker(dur),
		TimeOutDuration: dur,
		State:           Follower,
	}
}

func (s *Server) ResetTicker() {
	s.Timeout = time.NewTicker(s.TimeOutDuration)
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
	fmt.Println(s.Name+" Gave vote to ", request.Name)

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
			Name:         s.Name,
			Term:         s.CurrentTerm,
			CandidateId:  s.Id,
			LastLogIndex: len(s.Log) + 1,
			LastLogTerm:  s.Log[len(s.Log)-1],
		}
	}
	return RequestVote{
		Name:         s.Name,
		Term:         s.CurrentTerm,
		CandidateId:  s.Id,
		LastLogIndex: len(s.Log) + 1,
		LastLogTerm:  -1,
	}
}

func (s *Server) SendElection(sc Contact, rv RequestVote) []Response {
	return sc.SendRequestVote(rv)
}

func (s *Server) PromoteLeader(sc Count, responses ...Response) bool {
	total := sc.TotalServerCount()
	totalVotes := 0
	for _, response := range responses {
		if response.VoteGranted {
			totalVotes++
		}
	}

	if totalVotes > total/2 {
		s.State = Leader
		fmt.Println("Leader: ", s.Name)
		return true
	}
	return false
}

func (s *Server) ReceiveLog(log AppendLog) Response {
	if log.Term >= s.CurrentTerm {
		if log.Term != s.CurrentTerm {
			s.CurrentTerm = log.Term
		}
		// HeartBeat
		if len(log.Entries) == 0 {
			s.ResetTicker()
			return Response{
				Term:        s.CurrentTerm,
				VoteGranted: true,
			}

		}
	}
	return Response{
		Term:        s.CurrentTerm,
		VoteGranted: false,
	}

}

func (s *Server) SendLog(sc Contact, log AppendLog) []Response {
	return sc.SendAppendLog(log)
}

type Count interface {
	TotalServerCount() int
}

type Contact interface {
	SendRequestVote(vote RequestVote) []Response
	SendAppendLog(log AppendLog) []Response
}
