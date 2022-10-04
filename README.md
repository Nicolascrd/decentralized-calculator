#Decentralized Calculator

Decentralized calculator is a project to implement basic consensus algorithms.
Running the shell script will start 3 (for now) containers, each one implementing a basic calculator server.
The client can query any of the three containers, at :8001, :8002 or :8003
The consensus algorithm will then run with the containers and the client should receive its response, with one container "allowed" to failed.

## Context

This project is part of my end-of-studies research project on consensus algorithms.
