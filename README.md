# ZeroMqDemo
Go ZeroMQ examples

## Build and Run


```yaml
go install

./GO_PATH/bin/main <demo name>

or

cd PROJECT_DIR/main

go build

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
