markdown
# Project README

## Project Summary

This project is a work in progress. The goal is to build a system using Go, with services orchestrated through Docker Compose. Eventually, I plan to explore Kubernetes to run this project across multiple clusters. Right now, it’s a learning playground where I’m experimenting and improving my skills.

## Running the System

You can get the system up and running in two simple steps:

1. Start supporting services using Docker Compose  
   From the root directory, run:  
   ```bash
   docker compose -f docker-compose.yaml up
````

2. Run the Go application
   From the project root, execute:

   ```bash
   go run cmd/main.go
   ```

## Tech Stack

* Go (Golang) – main application
* Docker Compose – local service orchestration (`docker-compose.yaml`)
* Kubernetes (planned) – multi-cluster deployment

## Prerequisites

Before running the project, make sure you have:

* Go installed
* Docker installed
* Docker Compose (usually comes with Docker Desktop)
* Basic familiarity with terminal commands

## Project Status

The project is actively in progress. I’m currently learning Kubernetes to manage multiple clusters. Once I’m more confident, I’ll come back to integrate it fully into this system.
