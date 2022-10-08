# Decentralized Calculator

Decentralized calculator is a project to implement basic consensus algorithms.

Running the shell script will start some containers (3 by default), each one implementing a basic calculator server.

The client can query any of the containers, at :8001 for container 1, :8002 for container 2 ...

The consensus algorithm will then run with the containers and the client should receive its response, with one container "allowed" to failed.

## Context

This project is part of my end-of-studies research project on consensus algorithms.

## Testing

Launch the containers with :

```
sh launch.sh <3 4 5 6> # number of containers to start (min 3 max 9)

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