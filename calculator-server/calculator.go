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
	if !supported(parsed.OperationType) {
		http.Error(w, fmt.Sprintf("String %q passed to calculator does not represent a supported operation", parsed.OperationType), http.StatusBadRequest)
		return
	}

	// to implement

	res, _ = calculator(parsed.A, parsed.B, parsed.OperationType)

	io.WriteString(w, fmt.Sprint(res))
	return
}

func calculator(a int, b int, op string) (int, error) {
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		return a / b, nil
	}
	return 0, fmt.Errorf("String %q passed to calculator does not represent a supported operation", op)
}

func failingCalculator() int {
	return rand.Int()
}

func supported(op string) bool {
	switch op {
	case "+":
		return true
	case "-":
		return true
	case "*":
		return true
	case "/":
		return true
	}
	return false
}
