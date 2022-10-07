package main

import (
	"encoding/json"
	"fmt"
	"io"
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
	switch parsed.OperationType {
	case 1:
		res = parsed.A + parsed.B
	case 2:
		res = parsed.A - parsed.B
	case 3:
		res = parsed.A * parsed.B
	case 4:
		res = parsed.A / parsed.B
	}
	io.WriteString(w, fmt.Sprint(res))
	return
}
