package main

import (
	"os"
	"os/signal"
	server "raft-go"
	"syscall"
)

func main() {
	s1 := server.NewServer("A")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	sc := new(ServerHelper)
	s2 := server.NewServer("B")
	s3 := server.NewServer("C")
	sc.AddServerCount(&s2)
	sc.AddServerCount(&s3)
	dc := make(chan bool)
	go RunServer(&s1, sc, dc)
	go RunServer(&s2, sc, dc)
	go RunServer(&s3, sc, dc)
main:
	for {
		select {
		case <-sig:
			dc <- true
			break main
		}
	}
}

func RunServer(s *server.Server, sh *ServerHelper, doneChan chan bool) {
run:
	for {
		select {
		case <-doneChan:
			break run
		case <-s.Timeout.C:
			if s.State == server.Follower {
				if s.VotedFor == 0 {
					rv := s.StartElection()
					responses := s.SendElection(sh, rv)
					s.PromoteLeader(sh, responses...)
					hb := server.AppendLog{
						Term:         s.CurrentTerm,
						LeaderId:     s.Id,
						PrevLogIndex: s.CommitIndex,
						PrevLogTerm:  s.LastApplied,
						Entries:      []int{},
						LeaderCommit: 0,
					}
					s.SendLog(sh, hb)
				}
			} else if s.State == server.Leader {

			}
		}
	}
}

type ServerHelper struct {
	Servers []*server.Server
}

func (sc *ServerHelper) TotalServerCount() int {
	return len(sc.Servers)
}

func (sc *ServerHelper) AddServerCount(server *server.Server) {
	sc.Servers = append(sc.Servers, server)
}

func (sc *ServerHelper) RemoveServer(index int) {
	if sc.TotalServerCount() > index && index >= 0 {
		sc.Servers[index] = sc.Servers[len(sc.Servers)-1]
		sc.Servers = sc.Servers[:len(sc.Servers)-1]
	}
}

func (sc *ServerHelper) SendRequestVote(request server.RequestVote) (VoteResponse []server.Response) {
	for _, s := range sc.Servers {
		VoteResponse = append(VoteResponse, s.GiveElectionVote(request))
	}
	return VoteResponse
}

func (sc *ServerHelper) SendAppendLog(log server.AppendLog) (CommitResponse []server.Response) {
	for _, s := range sc.Servers {
		CommitResponse = append(CommitResponse, s.ReceiveLog(log))
	}
	return CommitResponse
}
