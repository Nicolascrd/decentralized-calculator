package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
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
				calc.leaderSendsHB()
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
	for id, addr := range calc.sys.addresses {
		if id == calc.ID {
			continue
		}
		wg.Add(1)
		go func(i int, addr string) {
			defer wg.Done()
			if calc.requestVote(addr) {
				numberOfVotes++
			}
		}(id, addr)
	}
	wg.Wait()
	if numberOfVotes <= calc.sys.numberOfNodes/2 { // no strict majority
		// not a leader
		calc.status = 2
		return
	}
	calc.leaderSendsHB()
}

func (calc *calculatorServer) leaderSendsHB() {
	// new leader
	calc.status = 1
	calc.leaderAddr = calc.addr
	calc.leaderID = calc.ID
	var wg sync.WaitGroup
	numOfValidations := 0
	doFollow := make([]int, 0)
	for id, addr := range calc.sys.addresses {
		if id == calc.ID {
			continue
		}
		wg.Add(1)
		go func(i int, addr string, df *[]int) {
			defer wg.Done()
			if calc.sendHB(addr) {
				numOfValidations++
				*df = append(*df, i)
			}
		}(id, addr, &doFollow)
	}
	wg.Wait()
	if numOfValidations < calc.sys.numberOfNodes-1 {
		calc.logger.Printf("Leader send HB process terminated, %d nodes do not follow : %v", calc.sys.numberOfNodes-1-numOfValidations, doFollow)
		calc.newSys(doFollow)
	}
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
		calc.logger.Printf("ERROR SENDING HB at %s : %s", addr, err.Error())
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
	resp, err := postJSON(calc.sys.addresses[node]+calculationInternalEndpoint, content, &calc.logger)

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

func (calc *calculatorServer) newSys(doFollow []int) {
	fmt.Printf("new sys called with doFollow %v and addresses %v", doFollow, calc.sys.addresses)
	sort.Slice(doFollow, func(i, j int) bool {
		return i < j
	})
	// doFollow contains the ids of the following nodes, sorted

	// self-update
	nbFollowers := len(doFollow)
	calc.sys.numberOfNodes = nbFollowers + 1
	newAddresses := make(map[int]string)
	newAddresses[calc.ID] = calc.addr
	for i := 0; i < nbFollowers; i++ {
		// calc.sys.addresses[doFollow[i]] is the address of the node (or addr for the leader)
		newAddresses[doFollow[i]] = calc.sys.addresses[doFollow[i]]
	}
	calc.sys.addresses = newAddresses
	calc.logger.Printf("New addresses list in system : %v", newAddresses)
}
