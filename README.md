# ZeroMqDemo
Go ZeroMQ examples based on https://github.com/pebbe/zmq4

## Build and Run


```yaml
go install

./GO_PATH/bin/main <demo name>

or

cd PROJECT_DIR/main

go build

list of demos: ./main --help

./main <demo name>

```

## Demos

- hwserver - dummy zeromq server

```yaml
   ./main hws
```
- hwclient - dummy zeromq client for hwserver

 ```yaml
    ./main hwc
```
- version - show ZeroMQ version
 ```yaml
    ./main -v
```

- wuserver - weather update pub-sub zeromq server

```yaml
   ./main wus
```
- wuclient - weather update pub-sub zeromq client for hwserver

 ```yaml
    ./main wuc <zipcode>
```
- taskVentilator - generator of parallel tasks which sends batch of tasks via socket

```yaml
   ./main tv
```

- taskWorker - worker which pulls messages from taskVentilator,does work and pushes to sink
```yaml
   ./main tw
```

- taskSink - sink pulling results from workers
```yaml
   ./main ts
```

- msreader - multiple socket reader by priority (requires running taskVentilator and wuserver)
```yaml
   ./main msr
```

- mspoller - multiple parallel socket reader (requires running taskVentilator and wuserver)
```yaml
   ./main msp
```

- rrclient - client which sends requests to rrworker through rrbroker and gets responses
```yaml
   ./main rrc
```

- rrbroker - broker for binding multiple rrclients and rrworkers. only one broker is required
```yaml
   ./main rrb
```

- rrworker - server which gets requests from  rrclient and sends response back
```yaml
   ./main rrw
```

- msgqueue - proxy based variant of rrbroker
```yaml
   ./main msgq
```

- interrupt - client for hwserver for proper CTRL-C handling via channels
```yaml
   ./main intс
```

- mtserver - multithreaded version of hwserver
```yaml
   ./main mts
```

- mtrelay - multithreaded relay
```yaml
   ./main mtr
```

- syncpub - synchronized publisher
```yaml
   ./main spb
```

- syncsub - synchronized subscriber
```yaml
   ./main ssb
```

- envpub - publisher with key in envelope
```yaml
   ./main epb
```

- envsub - subscriber  with key in envelope
```yaml
   ./main esb
```

- identity - demo showing different identitites for request reply pattern
```yaml
   ./main id
```

## The Load-Balancing Pattern

- routerReq - demo showing router-to-request pattern
```yaml
   ./main rtreq
```

- loadBalancingBroker - load-balancing broker demo  with embedded worker and client and using 0MQ high-level api for sending and receinving messages
```yaml
   ./main llb
```

- loadBalancingBrokerReactor - load-balancing broker using reactor
```yaml
   ./main llbr
```
====END PATTERN====

- asyncServer - asyncronious server with emedded client and embedded async workers
```yaml
   ./main asys
```

- peering - multibroker peering with embedded client and worker
```yaml
   ./main peer <broker> <peeers list>
```
## Lazy Pirate Pattern

Simple reliable request-reply pattern 

- lserver - server simulating issues
```yaml
   ./main ls
  ```

  - lclient - reliable client holding reconnection to lserver
  ```yaml
     ./main lc
  ```

## Simple Pirate Pattern
Basic reliable queing for Lazy Pirate

  - rqueue - reliable queue connecting lclient and rworker
  ```yaml
       ./main rq
  ```

  - rworker - reliable worker
  ```yaml
       ./main rw
  ```
## Paranoid Pirate Pattern
Robust reliable queuing

  - rqueue - robust reliable queue with heartbeat connecting lclient and rorworker
  ```yaml
       ./main rrq
  ```

  - rorworker - robust reliable worker
  ```yaml
       ./main rrw
  ```
### [Majordomo Protocol](https://rfc.zeromq.org/spec:7/MDP/) 
Service oriented reliable pattern
  - mdworker - worker 
  ```yaml
       ./main mdwr <true/false>
  ```
  - mdbroker - broker 
    ```yaml
         ./main mdbr <true/false>
    ```
  - mdclient - client 
     ```yaml
          ./main mdcl <true/false>
     ``` 
    
  - mdsrch - worker search discovery client. requires mdbroker and worker 
       ```yaml
            ./main mdsrch <true/false>
       ``` 
## Titanic
Disconnected reliable pattern based on Majordomo
(there is an issue with async mdapi)
- drbroker - broker 
    ```yaml
         ./main drbr <true/false>
    ```
 - drclient - client 
     ```yaml
          ./main drcl <true/false>
     ``` 
 Required mdbroker and mdworker. Starting sequence: mdworker,mdbroker, drbroker,drclient  
    
  
## Binary Star
Primary/backup high-availability pair

- hsrv - highly-available server. if true is set starts as primary server. 
   ```yaml
      ./main hsrv <true> 
    ``` 
     
- hclt - hiighly-available  client.

Two servers must be run: one in primary mode, another in backup 
Idea - stop primary server and restart it. The servers will change roles.   

## [Freelance Protocol](http://rfc.zeromq.org/spec:10.)

Brokerless connection between server and client

### Model 1: simple retry and failover

- brlServerFail - brokerless failover server
```yaml
./main bsf <endpoint>
```

- brlClientFail - brokerless failover client
```yaml
./main bcf <endpoint> <endpoint>
```

Run: start serveral servers and put as endpoints param to client servers addresses
### Model 2: send message from client, wait , try another server
- brlMsgServerFail - brokerless failover server receiving messages from client
```yaml
./main bsmf <endpoint>
```

- brlRepClientFail - brokerless failover client sending multiple replies to servers and showing avg roundtrip cost
```yaml
./main bcrf <endpoint> <endpoint>
```
Run: start serveral servers and put as endpoints param to client servers addresses

### Model 3: routing usage

- brlRtServer - brokerless routing server
```yaml
./main bsrtf <port> <true>
```

- brlRtClient - brokerless async routing client showing avg roundtrip cost
```yaml
./main bcrtf
```
Run: start two servers with ports 5555 and 5556 and one client.

## Espresso Pattern

Pub-Sub messages tracing

- pstracing - tracing messages comming from publisher to subscriber via pipe
```yaml
./main pst
```

## Last Value Caching

Model where new subscriber catches missed messages after joining

- ptlPub - publisher sending 1000 topics and one random update per second
```yaml
./main ptp <cacheEnabled=true/false>
```

- ptlSub - subscriber which conencts to one random topic and gets messages
```yaml
./main pts <cacheEnabled=true/false>
```

- lvcProxy - Last value caching proxy for data resending
```yaml
./main lvc 
```

Two subscriber required for demo. Runing order: lvc,pub, more than one sub.

## Suicidal Snail Pattern

Slow subscriber detection

- slowSubDetection - subscirber dies if msg is 1 second late. publisher send messsage with time stamp every msec
```yaml
./main ssd 
```

## Clone Pattern

Reliable Pub-Sub

- relPSServer - server
```yaml
./main rpss <primary=true/false>
```
- relPSClient - client
```yaml
./main rpsc 
```

Run: two servers(one primary,one backup) and two client

- fileTransfer - file transfer via ZeroMQ

```yaml
  ./main ft <filePath>
```