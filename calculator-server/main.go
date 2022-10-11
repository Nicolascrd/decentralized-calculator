package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const calculationEndpoint string = "/calc" // for client to query nodes

type system struct {
	NumberOfNodes int            `json:"numberOfNodes"` // number of nodes this node is linked to
	Addresses     map[int]string `json:"addresses"`     // ports of all the nodes this node is connected to
}

type calculatorServer struct {
	logger  log.Logger // associated logger
	addr    string     // URL in container eg centra-calcu-1:8000
	ID      int        // server number e.g. 1
	failing bool       // byzantine failure assumed to make the result of calculation random
	timeout <-chan time.Time
	sys     system // each node knows the system
}

type calculatorRequest struct {
	OperationType string `json:"operationType"` // + : add, - : substract, * : multiply or / : divide (euclidean)
	A             int    `json:"a"`             // first element
	B             int    `json:"b"`             // second element
}

type config struct {
	ConfigGraph [][]int `json:"configGraph"`
}

var globalConfig config

func main() {
	fmt.Println("Hello snowball calculator")
	args := os.Args[1:]
	if len(args) != 3 {
		fmt.Println("Wrong number of arguments in command line, expecting only 2 numbers between 0 and 99 and one bool")
		return
	}

	ind, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("First argument provided should be an int but \n " + err.Error())
		return
	}
	if ind < 0 || ind > 99 {
		fmt.Println("First Number given is out of bounds ([0,99])")
		return
	}
	tot, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("Second argument provided should be an int but \n" + err.Error())
		return
	}
	if tot < 0 || tot > 99 {
		fmt.Println("Second Number given is out of bounds ([0,99])")
		return
	}
	byz, err := strconv.ParseBool(args[2])
	if err != nil {
		fmt.Println("Third argument given should be an bool but \n " + err.Error())
		return
	}
	configFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println("Could not open config json : " + err.Error())
		return
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&globalConfig)
	if err != nil {
		fmt.Println("Could not decode config json : " + err.Error())
		return
	}

	configFile.Close()

	if len(globalConfig.ConfigGraph) == 0 {
		fmt.Println("Empty config graph")
		return
	}
	if len(globalConfig.ConfigGraph[0]) != len(globalConfig.ConfigGraph) {
		fmt.Println("Invalid config graph")
		return
	}
	if len(globalConfig.ConfigGraph) != tot {
		fmt.Printf("Config graph (%d nodes) does not correspond to the CL number of nodes (%d)", len(globalConfig.ConfigGraph), tot)
	}

	fmt.Println("config : ", globalConfig)

	calc := newCalculatorServer(ind, tot, byz)
	// go calc.launchTicker() // initiate timeouts --> start consensus process

	calc.launchCalculatorServer()
}

func newCalculatorServer(num int, tot int, failing bool) *calculatorServer {
	// num : number of this container (this node)
	// tot : total number of containers (nodes in the system)
	l := log.New(log.Writer(), "CalculatorServer - "+fmt.Sprint(num)+"  ", log.Ltime)
	c := make(chan time.Time)

	addresses := make(map[int]string)
	for i := 1; i <= tot; i++ {
		addresses[i] = "decentra-calcu-" + fmt.Sprint(i) + ":8000"
	}
	sys := system{
		NumberOfNodes: tot,
		Addresses:     addresses,
	}

	return &calculatorServer{
		logger:  *l,
		ID:      num,
		addr:    "decentra-calcu-" + fmt.Sprint(num) + ":8000",
		timeout: c,
		sys:     sys,
		failing: failing,
	}
}

func (calc *calculatorServer) launchCalculatorServer() {
	http.HandleFunc(calculationEndpoint, calc.calcHandler)

	err := http.ListenAndServe(calc.addr, nil)
	if err != nil {
		calc.logger.Panicln("Cannot launch server")
	}
}
