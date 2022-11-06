# websocket-benchmark-cli

- 支持连接压测
- 支持IO吞吐压测

## websocket-benchmark-cli
```bash
websocket-benchmark-cli -h
NAME:
   websocket-benchmark-cli - websocket benchmark tool

USAGE:
   websocket-benchmark-cli [global options] command [command options] [arguments...]

VERSION:
   v1.0.1

COMMANDS:
   conn     websocket connection benchmark
   message  websocket message benchmark
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --host value             (default: "localhost:8080")
   --path value             (default: "/ws")
   --user value             (default: 500)
   --connectInterval value  (default: "20ms")
   --help, -h               show help (default: false)
   --version, -v            print the version (default: false)
```

### 全局参数
- host: websocket服务host
- path: websocket路径
- user: 用户连接数量
- connectInterval: 连接间隔

## 连接压测

#### command
```bash
websocket-benchmark-cli conn -h
NAME:
   websocket-benchmark-cli conn - websocket connection benchmark

USAGE:
   websocket-benchmark-cli conn [command options] [arguments...]

OPTIONS:
   --host value             (default: "localhost:8080")
   --path value             (default: "/ws")
   --user value             (default: 500)
   --connectInterval value  (default: "20ms")
   
   --sleep value            (default: "60m")
   --help, -h               show help (default: false)
```

- sleep：连接保持时间

#### example
```
websocket-benchmark-cli conn --host localhost:8888 --connectInterval 30ms --user 10000 --sleep 1h
```

## 消息吞吐压测

#### command
```bash
websocket-benchmark-cli message -h
NAME:
   websocket-benchmark-cli message - websocket message benchmark

USAGE:
   websocket-benchmark-cli message [command options] [arguments...]

OPTIONS:
   --host value             (default: "localhost:8080")
   --path value             (default: "/ws")
   --user value             (default: 500)
   --connectInterval value  (default: "20ms")
   
   --content value          (default: "hello world")
   --file value             
   --interval value         (default: "1s")
   --times value            (default: 100)
   --help, -h               show help (default: false)
```

- content：发送的消息主体
- file：读取的文件作为发送的消息主体，适合大文件
- interval：发送消息的间隔
- times：发送消息的总次数

#### example
```bash
websocket-benchmark-cli message --host localhost:8888 --connectInterval 30ms --user 1000 --content messagebody
```


### todo
- report指标