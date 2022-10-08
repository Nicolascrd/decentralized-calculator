package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const voteEndpoint string = "/vote"
const heartbeatEndpoint string = "/heartBeat"
const calculationEndpoint string = "/calc"                  // for client to query nodes
const calculationInternalEndpoint string = "/calc-internal" // for leader to query nodes

type system struct {
	numberOfNodes int            // number of nodes in the whole system
	addresses     map[int]string // ports of all nodes in order (including this one)
}

type calculatorServer struct {
	logger      log.Logger // associated logger
	addr        string     // URL in container eg centra-calcu-1:8000
	ID          int        // server number e.g. 1
	leaderID    int        // server number corresponding to known leader
	leaderAddr  string     // port associated to leader
	status      int        // 1 for leader, 2 for follower, 3 for candidate not sure I should keep that
	hbReceived  bool       // true if a heart beat from the leader was received, reset to false at each tick from ticker
	currentTerm int        // term is the period
	votedFor    int        // id of node the server voted for in the current term
	timeout     <-chan time.Time
	sys         system // each node knows the system
}

type calculatorRequest struct {
	OperationType int `json:"operationType"` // 1 : add,2 : substract,3 : multiply or 4 : divide
	A             int `json:"a"`             // first element
	B             int `json:"b"`             // second element
}

func main() {
	fmt.Println("Hello calculator")
	args := os.Args[1:]
	if len(args) != 2 {
		fmt.Println("Wrong number of arguments in command line, expecting only two number between 0 and 9")
		return
	}

	ind, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("First argument provided should be an int but \n " + err.Error())
		return
	}
	if ind < 0 || ind > 9 {
		fmt.Println("First Number given is out of bounds ([0,9])")
		return
	}
	tot, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("Second argument provided should be an int but \n" + err.Error())
		return
	}
	if tot < 0 || tot > 9 {
		fmt.Println("Second Number given is out of bounds ([0,9])")
		return
	}
	calc := newCalculatorServer(ind, tot)
	go calc.launchTicker() // initiate timeouts

	calc.launchCalculatorServer()
}

func newCalculatorServer(num int, tot int) *calculatorServer {
	// num : number of this container (this node)
	// tot : total number of containers (nodes in the system)
	l := log.New(log.Writer(), "CalculatorServer - "+fmt.Sprint(num)+"  ", log.Ltime)
	c := make(chan time.Time)

	addresses := make(map[int]string)
	for i := 1; i <= tot; i++ {
		addresses[i] = "decentra-calcu-" + fmt.Sprint(i) + ":8000"
	}
	sys := system{
		numberOfNodes: tot,
		addresses:     addresses,
	}

	return &calculatorServer{
		logger:      *l,
		ID:          num,
		addr:        "decentra-calcu-" + fmt.Sprint(num) + ":8000",
		timeout:     c,
		sys:         sys,
		status:      2,
		currentTerm: 0, // currentTerm is incremented and starts at 1 at first apply
	}
}

func (calc *calculatorServer) launchCalculatorServer() {
	http.HandleFunc(calculationEndpoint, calc.calcHandler)
	http.HandleFunc(heartbeatEndpoint, calc.heartBeatHandler)
	http.HandleFunc(voteEndpoint, calc.vote)
	http.HandleFunc(calculationInternalEndpoint, calc.calcInternalHandler)

	err := http.ListenAndServe(calc.addr, nil)
	if err != nil {
		calc.logger.Panicln("Cannot launch server")
	}
}
