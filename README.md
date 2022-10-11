# Decentralized Calculator

Decentralized calculator is a project to implement basic consensus algorithms.

Running the shell script will start some containers (3 by default), each one implementing a basic calculator server.

Some containers can contain a failing calculator which replies a random integer instead of the calculation.
To tackle this, a parameter allows the user to switch the leader query to a majority vote.

The client can query any of the containers, at :8001 for container 1, :8002 for container 2 ...

The consensus algorithm will then run with the containers and the client should receive its response, with one container "allowed" to fail.

Master branch implements [the Raft consensus algorithm](https://raft.github.io/raft.pdf)

Snowball branch implements [the Snowball consensus algorithm](https://assets.website-files.com/5d80307810123f5ffbb34d6e/6009805681b416f34dcae012_Avalanche%20Consensus%20Whitepaper.pdf)

## Context

This project is part of my end-of-studies research project on consensus algorithms.

Check out my [bibliographic survey on consensus algorithms](https://github.com/Nicolascrd/researchProjectConsensus/blob/master/biblio-rp.pdf)

## Testing

Select your parameters in config.json :

```
{
    "updateSystem" : If true the systems remap when one node is failing completely (failing to reply to request within 200ms)
    "majorityVoteCalculation" : If true the leader triggers a majority vote for each calculation (all nodes calculate and vote)
}
```


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
    -d '{"operationType":"-","a":10, "b":2}' # ask container <1 2 3> for the result of 10 - 2
```
Supported operations are +, -, *, / (euclidian division)
\
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
