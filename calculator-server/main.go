package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

type calculatorServer struct {
	logger log.Logger // associated logger
	addr   string     // port e.g. :8001
	number int        // server number e.g. 1
}

type calculatorRequest struct {
	OperationType int `json:"operationType"` // 1 : add,2 : substract,3 : multiply or 4 : divide
	A             int `json:"a"`             // first element
	B             int `json:"b"`             // second element
}

func main() {
	fmt.Println("Hello calculator")
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println("Too many (or no) arguments in command line, expecting only one number between 0 and 9")
		return
	}

	ind, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Argument provided should be an int but \n " + err.Error())
		return
	}
	if ind < 0 || ind > 9 {
		fmt.Println("Number given is out of bounds ([0,9])")
		return
	}

	calc := newCalculatorServer(ind, ":800"+args[0])

	calc.launchCalculatorServer()
}

func newCalculatorServer(num int, str string) *calculatorServer {
	l := log.New(log.Writer(), "CalculatorServer - "+fmt.Sprint(num)+"  ", log.Ltime)
	return &calculatorServer{
		logger: *l,
		number: num,
		addr:   str,
	}
}

func (calc *calculatorServer) launchCalculatorServer() {
	http.HandleFunc("/", calc.handler)

	err := http.ListenAndServe(calc.addr, nil)
	if err != nil {
		calc.logger.Panicln("Cannot launch server")
	}

}

func (calc *calculatorServer) handler(w http.ResponseWriter, r *http.Request) {
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
