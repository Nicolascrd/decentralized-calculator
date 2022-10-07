package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
)

func (calc *calculatorServer) calcHandler(w http.ResponseWriter, r *http.Request) {
	var parsed calculatorRequest
	var res int
	err := json.NewDecoder(r.Body).Decode(&parsed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	calc.logger.Println("Receive calculation :", parsed)
	if calc.status != 1 {
		// if calc is posted to a node which is not the leader
		res = calc.transferLeader(parsed)
	} else {
		// replace --> ask a random follower or to the leader himself

		res = calc.transferFromLeader(rand.Intn(calc.sys.numberOfNodes)+1, parsed)
	}
	io.WriteString(w, fmt.Sprint(res))
	return
}

func (calc *calculatorServer) calcInternalHandler(w http.ResponseWriter, r *http.Request) {
	var parsed calculatorRequest
	var res int
	err := json.NewDecoder(r.Body).Decode(&parsed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	calc.logger.Println("Receive calculation from leader :", parsed)
	res = calculator(parsed.A, parsed.B, parsed.OperationType)
	io.WriteString(w, fmt.Sprint(res))
	return
}

func calculator(a int, b int, op int) int {
	switch op {
	case 1:
		return a + b
	case 2:
		return a - b
	case 3:
		return a * b
	case 4:
		return a / b
	}
	return 0
}
