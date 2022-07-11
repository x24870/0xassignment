# assignment
---
## Pre-required
### Initialize Project
Init packages
```
$go mod tidy
```
Get PostgresSQL docker image
```
$docker pull postgres
```
Runs a database locally
```
$docker-compose up -d
```
### Configure Infura API endpoints 
Please copy your Infura API endpoints and paste to **local_dev/localrc**
```
export INFURA_ENDPOINT=YOUR_INFURA_ENDPOINT
export INFURA_WS_ENDPOINT=YOUR_INFURA_WS_ENDPOINT
```
---
## Run
```
make run
```
---
## Query API
### examples
```
curl http://127.0.0.1:8000/blocks?limit=20
curl http://127.0.0.1:8000/blocks/15118398
curl http://127.0.0.1:8000/transaction/0xf61c08a876e6c04aa24de03b381ffbf7bd36ca9fc0b19b4709f2b13867cf04f9
```
---
## System design

### DB

![alt text](https://github.com/x24870/0xassignment/blob/master/docs/db.jpg)

### Workflow

![alt text](https://raw.githubusercontent.com/x24870/0xassignment/master/docs/workflow.jpg)




