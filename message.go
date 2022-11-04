package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var CmdMessageFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "content",
		Value: "hello world",
	},
	&cli.StringFlag{
		Name: "file",
	},
	&cli.StringFlag{
		Name:  "interval",
		Value: "1s",
	},
	&cli.IntFlag{
		Name:  "times",
		Value: 100,
	},
}

var CmdMessage = &cli.Command{
	Name:   "message",
	Usage:  "websocket message benchmark",
	Flags:  append(GlobalFlags, CmdMessageFlags...),
	Action: CmdMessageAction,
}

func CmdMessageAction(ctx *cli.Context) error {
	endpoint := ctx.String("host")
	path := ctx.String("path")
	userNum := ctx.Int("user")
	connectIntv := ctx.Duration("connectInterval")

	messageContent := ctx.String("content")
	messageIntv := ctx.Duration("interval")
	messageTimes := ctx.Int("times")
	messageFile := ctx.String("file")

	if messageFile != "" {
		content, err := os.ReadFile(messageFile)
		if err != nil {
			panic(err)
		}
		messageContent = string(content)
	}

	benchmarker := NewWebsocketBenchmarker(
		WebsocketBenchmarkerOptionEndpoint(endpoint),
		WebsocketBenchmarkerOptionPath(path),
		WebsocketBenchmarkerOptionUserNum(userNum),
		WebsocketBenchmarkerOptionConnectInterval(connectIntv),
		WebsocketBenchmarkerOptionMessage(messageContent),
		WebsocketBenchmarkerOptionMessageTimes(messageTimes),
		WebsocketBenchmarkerOptionMessageInterval(messageIntv),
	)

	log.Println(benchmarker.MessageInfo())

	c := make(chan os.Signal, 1)
	signal.Notify(c, []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT}...)

	errC := make(chan error, 1)
	go func() {
		if err := benchmarker.Test(); err != nil {
			errC <- err
		}
		if err := benchmarker.StartMessageBenchmark(); err != nil {
			errC <- err
		}
		errC <- nil
	}()
	select {
	case s := <-c:
		benchmarker.Stop()
		log.Printf("process stop by signal %s\n", s.String())
	case err := <-errC:
		benchmarker.Stop()
		if err != nil {
			log.Printf("process stop by benchmark err: %s\n", err.Error())
		} else {
			log.Println("process complete")
		}
	}
	return nil
}
