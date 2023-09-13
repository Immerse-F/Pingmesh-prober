
The project data has been anonymized.
# Paper
(PingMesh)[https://dl.acm.org/doi/pdf/10.1145/2829988.2787496]

# Scenario

Distributed multi-machine storage for large-scale network performance monitoring, meeting the requirement of generating 200 GB of data per day.

# Available Solutions
gRPC + PingMesh (Controller & Agent architecture) + Prometheus federation cluster + Grafana/other visualization technologies



# Golang Installation (Required for both Agent and Controller)
apt install golang-go


# Golang Proxy Installation (Required for both Agent and Controller)
export GO111MODULE=on
export GOPROXY=https://mirrors.aliyun.com/goproxy/ 

# Agent Deployment
The Agent needs to be deployed on all nodes that require monitoring.

Modify the XML file, and change the node_name to the corresponding node.


# Controller Deployment

The Controller can be deployed on multiple machines, and one Controller can manage more than 10 Agents.

# prober.yml 
Modify  ```rpc_listen_addr``` to the IP + port of the Controller.
Modify ```metrics_listen_addr``` to the same IP + port.

# pkg/cmd/agent/main.go
Modify the ```grpcServerAddress``` IP address in this file to the IP address of the Controller and the corresponding port.


# Prometheus安装

After installation, configure the IP and port in the respective Prometheus folder under /usr/local with the yml file.
Run 
```
./prometheus --config.file=prometheus.yml to start Prometheus.
```
# Start the Controller
```
go run pkg/cmd/server/main.go
```


# Start the Agent
```
go run pkg/cmd/agent/main.go
```


### Download Prometheus Data to Local
vim ./query.go 
#Edit query.go and modify the start and end timestamps.



### Additional Custom Operations under gRPC Communication Framework
#Proto communication
```
protoc --go_out=. --go-grpc_out=. prober.proto
```





