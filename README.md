# Decentralized Calculator

Decentralized calculator is a project to implement basic consensus algorithms.

Running the shell script will start some containers (3 by default), each one implementing a basic calculator server.

Some containers can contain a failing calculator which replies a random integer instead of the calculation.
To tackle this, a parameter allows the user to switch the leader query to a majority vote.

The client can query any of the containers, at :8001 for container 1, :8002 for container 2 ...

The consensus algorithm will then run with the containers and the client should receive its response, with one container "allowed" to fail.

Master branch implements [the Raft consensus algorithm](https://raft.github.io/raft.pdf)

Initially I wanted to have a Snowball branch which would implement [the Snowball consensus algorithm](https://assets.website-files.com/5d80307810123f5ffbb34d6e/6009805681b416f34dcae012_Avalanche%20Consensus%20Whitepaper.pdf). 

But I realised starting the implementation that the Snowball consensus algorithm is not at all a good fit for a decentralized calculator. 
That is for a few reasons :
- In a decentralized calculator, each node implements one calculator and the client queries a calculation. In Snowball, nodes don't have a preference for any value, they are just trying to reach consensus fast with the values the clients provides.
- I wanted to tweek the snowball algorithm to make the node calculate at some point to reach consensus on the answer value, not on the query. However if the nodes calculate when they are blank, should they return the value they calculated ( as if they were colored ) or should they ask a quorum for an anwer to propagate the request ? Should they do both ?
- In fact, if they return the value they calculated, there is no propagation, no majority and no consensus. If they ask a quorum for an answer, we would have to compute at some point (because otherwise they will never be a majority). Therefore we would need to track the depth and assume that at depth = d , the nodes start computing an answer, which for my example would be the answer of the calculation, or a random integer if the node is faulty. Then the actual consensus protocol can run, maybe the first node can launch it and so on, but we are getting away from the original protocol and adding a lot of complexity for no benefit.
- One option I thought could be a solution would be to simply send the query to every node from the first node. Each node does the computation. Then each node has a value, and we then run the Snowball consensus from the first node, with all nodes already colored. But this removes one of the benefit of snowball with is that all nodes are not supposed to be directly connected to each other.

I would be able to tweek the snowball consensus protocol and make it work. But the original idea of the protocol would be lost. 
Therefore I am starting a new project with a replicated state machine to compare both algorithms in a more interesting context.


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
