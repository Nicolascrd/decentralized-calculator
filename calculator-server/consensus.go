package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
)

type heartBeatRequest struct {
	LeaderID   int    `json:"leaderID"`
	LeaderAddr string `json:"leaderAddr"`
	LeaderTerm int    `json:"leaderTerm"`
}

type heatBeatResponse struct {
	CurrentTerm int  `json:"currentTerm"`
	Success     bool `json:"success"`
}

type voteRequest struct {
	CandidateID int `json:"candidateID"`
	Term        int `json:"term"`
}

type voteResponse struct {
	Term        int  `json:"term"`
	VoteGranted bool `json:"voteGranted"`
}

func (calc *calculatorServer) heartBeatHandler(w http.ResponseWriter, r *http.Request) {
	// new leader or old leader
	var parsed heartBeatRequest
	err := json.NewDecoder(r.Body).Decode(&parsed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	calc.logger.Printf("New heartbeat from %d, term %d", parsed.LeaderID, parsed.LeaderTerm)
	if parsed.LeaderTerm < calc.currentTerm {
		calc.logger.Printf("Heartbeat rejected: leaderTerm %d is lower than currentTerm %d", parsed.LeaderTerm, calc.currentTerm)
		replyJSON(w, heatBeatResponse{
			CurrentTerm: calc.currentTerm,
			Success:     false,
		}, &calc.logger)
		return
	}
	calc.hbReceived = true
	calc.leaderID = parsed.LeaderID
	calc.leaderAddr = parsed.LeaderAddr
	calc.currentTerm = parsed.LeaderTerm
	calc.status = 2 // switch to follower / stay as follower
	calc.logger.Printf("Heartbeat accepted")
	replyJSON(w, heatBeatResponse{
		CurrentTerm: calc.currentTerm,
		Success:     true,
	}, &calc.logger)
	return
}

func (calc *calculatorServer) vote(w http.ResponseWriter, r *http.Request) {
	// a candidate asks the node for a vote to become the leader
	var parsed voteRequest
	err := json.NewDecoder(r.Body).Decode(&parsed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	calc.logger.Printf("New Vote request from %d, term %d", parsed.CandidateID, parsed.Term)
	if parsed.Term < calc.currentTerm {
		calc.logger.Printf("Vote request rejected: candidateTerm %d is lower than currentTerm %d", parsed.Term, calc.currentTerm)
		replyJSON(w, voteResponse{
			Term:        calc.currentTerm,
			VoteGranted: false,
		}, &calc.logger)
		return
	}
	calc.logger.Printf("Vote request accepted")
	calc.votedFor = parsed.CandidateID
	calc.currentTerm = parsed.Term
	calc.status = 2 // switch to follower, stay as follower
	replyJSON(w, voteResponse{
		Term:        calc.currentTerm,
		VoteGranted: true,
	}, &calc.logger)
	return

}

func (calc *calculatorServer) launchTicker() {
	ticker := time.NewTicker(5 * time.Second)
	calc.timeout = ticker.C
	for {
		select {
		case <-ticker.C:
			if calc.status == 1 {
				// leader does not expect HB, but sends them
				calc.logger.Printf("Leader sends HB")
				calc.leaderSendHB()
			} else {
				if calc.hbReceived {
					calc.logger.Printf("Ticker ticked, with heartbeat received")
					calc.hbReceived = false
				} else {
					calc.logger.Printf("Ticker ticked, with heartbeat not received")
					calc.apply()
				}
			}
		}
	}
}

func (calc *calculatorServer) apply() {
	// apply to become the leader, change status immediately to candidate
	numberOfVotes := 1 //the server votes for itself
	calc.status = 3
	calc.currentTerm++
	var wg sync.WaitGroup
	for i := 1; i <= calc.sys.numberOfNodes; i++ {
		if i == calc.ID {
			continue
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if calc.requestVote(calc.sys.addresses[i-1]) {
				numberOfVotes++
			}
		}(i)
	}
	wg.Wait()
	if numberOfVotes <= calc.sys.numberOfNodes/2 { // no strict majority
		// not a leader
		calc.status = 2
		return
	}
	calc.leaderSendHB()
}

func (calc *calculatorServer) leaderSendHB() {
	// new leader
	calc.status = 1
	calc.leaderAddr = calc.addr
	calc.leaderID = calc.ID
	var wg sync.WaitGroup
	numOfValidations := 0
	for i := 1; i <= calc.sys.numberOfNodes; i++ {
		if i == calc.ID {
			continue
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if calc.sendHB(calc.sys.addresses[i-1]) {
				numOfValidations++
			}
		}(i)
	}
	wg.Wait()
	calc.logger.Printf("Leader send HB process terminated, %d nodes follow", numOfValidations)
}

func (calc *calculatorServer) requestVote(addr string) bool {
	var response voteResponse
	calc.logger.Printf("Request vote sent to %s", addr)
	resp, err := postJSON(addr+voteEndpoint, voteRequest{CandidateID: calc.ID, Term: calc.currentTerm}, &calc.logger)
	if err != nil {
		calc.logger.Printf("Error requesting vote at %s : %s", addr, err.Error())
		return false
	}

	res, err := decodeJSONResponse(resp, &calc.logger)

	err = mapstructure.Decode(res, &response)
	if err != nil {
		calc.logger.Printf("Error parsing vote from %s : %s", addr, err.Error())
		return false
	}

	return response.VoteGranted
}

func (calc *calculatorServer) sendHB(addr string) bool {
	var response heatBeatResponse
	resp, err := postJSON(addr+heartbeatEndpoint, heartBeatRequest{LeaderID: calc.ID, LeaderAddr: calc.addr, LeaderTerm: calc.currentTerm}, &calc.logger)

	if err != nil {
		calc.logger.Printf("Error sending HB at %s : %s", addr, err.Error())
		return false
	}

	res, err := decodeJSONResponse(resp, &calc.logger)

	if err != nil {
		calc.logger.Printf("Error decoding JSON HB response at %s : %s", addr, err.Error())
		return false
	}

	err = mapstructure.Decode(res, &response)
	if err != nil {
		calc.logger.Printf("Error parsing vote from %s : %s", addr, err.Error())
		return false
	}
	return response.Success
}

func (calc *calculatorServer) transferLeader(content calculatorRequest) int {
	resp, err := postJSON(calc.leaderAddr+calculationEndpoint, content, &calc.logger)

	if err != nil {
		calc.logger.Printf("Error transfering calculation to leader : %s", err.Error())
		return 0
	}

	integer, err := decodeIntResponse(resp, &calc.logger)

	if err != nil {
		calc.logger.Printf("Error decoding int response from leader : %s", err.Error())
		return 0
	}

	return integer
}

func (calc *calculatorServer) transferFromLeader(node int, content calculatorRequest) int {
	resp, err := postJSON(calc.sys.addresses[node-1]+calculationInternalEndpoint, content, &calc.logger)

	calc.logger.Printf("Transfering calculation to node n°%d", node)

	if err != nil {
		calc.logger.Printf("Error transfering calculation to leader : %s", err.Error())
		return 0
	}

	integer, err := decodeIntResponse(resp, &calc.logger)

	if err != nil {
		calc.logger.Printf("Error decoding int response from leader : %s", err.Error())
		return 0
	}

	return integer
}
