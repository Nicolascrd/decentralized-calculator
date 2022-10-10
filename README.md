# Decentralized Calculator

Decentralized calculator is a project to implement basic consensus algorithms.

Running the shell script will start some containers (3 by default), each one implementing a basic calculator server.

Some containers can contain a failing calculator, at the moment no check are made on the result and some times the random result will be returned.

The client can query any of the containers, at :8001 for container 1, :8002 for container 2 ...

The consensus algorithm will then run with the containers and the client should receive its response, with one container "allowed" to fail.

## Context

This project is part of my end-of-studies research project on consensus algorithms.

Check out my bibliographic survey on consensus algorithms at https://github.com/Nicolascrd/researchProjectConsensus/blob/master/biblio-rp.pdf 

## Testing

Launch the containers with :

```
sh launch.sh <3 4 5 6> <0 1 2>
# number of containers to start (min 3 max 99) and number of containers with a failing calculator
```
The first <0 1 2> containers will have a failing calculator

\
Query the containers with :

```
curl -X POST localhost:800<1 2 3>/calc -H 'Content-Type: application/json' \
    -d '{"operationType":2,"a":10, "b":2}' # ask container <1 2 3> for the result of 10 - 2
```
\
Check the logs of container <1 2 3> with :

```
docker logs --follow decentra-calcu-<1 2 3>
```
\
You can also try crashing containers with :

```
docker stop decentra-calcu-<1 2 3>
```
\
See what happens in the logs as the consensus algorithm run !
