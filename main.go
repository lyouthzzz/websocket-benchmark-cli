package main

import (
	"github.com/urfave/cli"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	app := cli.NewApp()

	app.Name = "wsbench"
	app.Usage = "websocket benchmark tool"
	app.Version = "v1.0.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "message", Value: "hello world"},
		cli.StringFlag{Name: "messageInterval", Value: "1s"},
		cli.IntFlag{Name: "messageTimes", Value: 60},
		cli.StringFlag{Name: "host", Value: "localhost:8080"},
		cli.StringFlag{Name: "path", Value: "/ws"},
		cli.IntFlag{Name: "user", Value: 500},
	}

	app.Action = func(cli *cli.Context) {
		message := cli.String("message")
		messageInterval := cli.String("messageInterval")
		messageTimes := cli.Int("messageTimes")
		endpoint := cli.String("host")
		path := cli.String("path")
		userNum := cli.Int("user")

		intv, err := time.ParseDuration(messageInterval)
		if err != nil {
			panic(err)
		}

		benchmarker := NewWebsocketBenchmarker(
			WebsocketBenchmarkerOptionEndpoint(endpoint),
			WebsocketBenchmarkerOptionMessage(message),
			WebsocketBenchmarkerOptionMessageTimes(messageTimes),
			WebsocketBenchmarkerOptionPath(path),
			WebsocketBenchmarkerOptionMessageInterval(intv),
			WebsocketBenchmarkerUserNum(userNum),
		)

		c := make(chan os.Signal, 1)
		signal.Notify(c, []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT}...)

		errC := make(chan error, 1)

		go func() {
			if err := benchmarker.Test(); err != nil {
				errC <- err
			}
			if err := benchmarker.Start(); err != nil {
				errC <- err
			}
			errC <- nil
		}()
		select {
		case s := <-c:
			benchmarker.Stop()
			time.Sleep(10 * time.Second)
			log.Printf("process down by signal %s\n", s.String())
		case err := <-errC:
			if err != nil {
				log.Printf("process down by benchmark err: %s\n", err.Error())
			} else {
				log.Println("process down")
			}
		}
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
