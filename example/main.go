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
main:
	for {
		select {
		case <-s1.Timeout.C:
			if s1.VotedFor == 0 {
				rv := s1.StartElection()
				responses := s1.SendElection(sc, rv)
				s1.PromoteLeader(sc, responses...)
				break main
			}
		case <-sig:
			break main
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

func (sc *ServerHelper) SendRequestVote(request server.RequestVote) (VoteResponse []server.VoteResponse) {
	for _, s := range sc.Servers {
		VoteResponse = append(VoteResponse, s.GiveElectionVote(request))
	}
	return VoteResponse
}
