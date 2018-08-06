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
   ./main hwserver
```
- hwclient - dummy zeromq client for hwserver

 ```yaml
    ./main hwclient
```
- wuserver - weather update pub-sub zeromq server

```yaml
   ./main wuserver
```
- wuclient - weather update pub-sub zeromq client for hwserver

 ```yaml
    ./main wuclient <zipcode>
```
- taskVentilator - generator of parallel tasks which sends batch of tasks via socket

```yaml
   ./main taskVentilator
```

- taskWorker - worker which pulls messages from taskVentilator,does work and pushes to sink
```yaml
   ./main taskWorker
```

- taskSink - sink pulling results from workers
```yaml
   ./main taskSink
```

- msreader - multiple socket reader by priority (requires running taskVentilator and wuserver)
```yaml
   ./main msreader
```

- mspoller - multiple parallel socket reader (requires running taskVentilator and wuserver)
```yaml
   ./main mspoller
```

- rrclient - client which sends requests to rrworker through rrbroker and gets responses
```yaml
   ./main rrclient
```

- rrbroker - broker for binding multiple rrclients and rrworkers. only one broker is required
```yaml
   ./main rrbroker
```

- rrworker - server which gets requests from  rrclient and sends response back
```yaml
   ./main rrworker
```

- msgqueue - proxy based variant of rrbroker
```yaml
   ./main msgqueue
```

- interrupt - client for hwserver for proper CTRL-C handling via channels
```yaml
   ./main interrupt
```

- mtserver - multithreaded version of hwserver
```yaml
   ./main mtserver
```

- mtrelay - multithreaded relay
```yaml
   ./main mtrelay
```

- syncpub - synchronized publisher
```yaml
   ./main syncpub
```

- syncsub - synchronized subscriber
```yaml
   ./main syncsub
```

- envpub - publisher with key in envelope
```yaml
   ./main envpub
```

- envsub - subscriber  with key in envelope
```yaml
   ./main envsub
```

- identity - demo showing different identitites for request reply pattern
```yaml
   ./main identity
```

- routerReq - demo showing router-to-request pattern
```yaml
   ./main rtreq
```

- loadBalancingBroker - load-balancing broker demo  with embedded worker and client and using 0MQ high-level api for sending and receinving messages
```yaml
   ./main llbroker
```

- loadBalancingBrokerReactor - load-balancing broker using reactor
```yaml
   ./main llbroker
```

- ayncServer - asyncronious server with emedded client and embedded async workers
```yaml
   ./main asyncsrv
```

- peering - multibroker peering with embedded client and worker
```yaml
   ./main peering <broker> <peeers list>
```

- lserver - server simulating issues
```yaml
   ./main ls
  ```

  - lclient - reliable client holding reconnection to lserver
  ```yaml
     ./main lc
    ```

  - rqueue - reliable queue connecting lclient and rworker
  ```yaml
       ./main rq
  ```

  - rworker - reliable worker
  ```yaml
       ./main rw
  ```

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