### Testing multiple calculator
sh launch.sh

curl -X POST localhost:800<1 2 3> -H 'Content-Type: application/json' -d '{"operationType":2,"a":10, "b":2}'

### Testing one calculator
go build

./calculator-server 1

curl -X POST localhost:8001 -H 'Content-Type: application/json' -d '{"operationType":2,"a":10, "b":2}'

ans --> 8 (10-2)