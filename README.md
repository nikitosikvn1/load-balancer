# SEC-Lab-4

This repository contains the completed tasks for the software engineering lab assignment focused on horizontal scaling of software systems. The tasks involved implementing load balancing algorithms, integration testing, and setting up continuous integration.

## Contributors:
- Danil Yaremenko danilyaremenko@gmail.com
- Nikita Petrykin n.petrykin.im12@kpi.ua
- Yurii Grygorash gyv220427@gmail.com
- Yan Petrov yanemerald2004@gmail.com

## Overview

The goal of this lab assignment was to deepen understanding of load balancing principles and implementation, and to practice integration testing. The tasks were carried out using a repository template provided in the assignment, which included necessary files and instructions for setting up the environment using Docker and docker-compose.

## Implemented features

### 1. Balancing algorithm:
djb2 hashes the url path where the request is sent to:
```go
pathHash := hash(r.URL.Path)
serverIndex := int(pathHash) % len(healthyServers)
```
### 2. Unit tests: 
Tests that check each balancer components work separately. 

### 3. Integration tests: 
Tests wich check that fully prepared balancer works as expected.

## Running the Project

The final step of the lab assignment was to build and run the project using Docker. To do this, follow the steps below:

1. Ensure Docker and docker-compose are installed on your machine.

2. Clone this repository to your local machine.

3. Navigate to the project directory in your terminal.

4. Run the following command to build and run the project:

```bash
docker-compose up --build
```

## Running tests

To run the tests, run the following command:

```bash
go test -v ./...
```

## Running integration tests

To run the integration tests, run the following command:

```bash
docker-compose -f docker-compose.yaml -f docker-compose.test.yaml \ up --exit-code-from test
```