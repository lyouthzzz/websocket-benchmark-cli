package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var CmdConnFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "sleep",
		Value: "60m",
	},
}

var CmdConn = &cli.Command{
	Name:   "conn",
	Usage:  "websocket connection benchmark",
	Flags:  append(GlobalFlags, CmdConnFlags...),
	Action: CmdConnRunAction,
}

func CmdConnRunAction(ctx *cli.Context) error {
	endpoint := ctx.String("host")
	path := ctx.String("path")
	userNum := ctx.Int("user")
	connectInterval := ctx.Duration("connectInterval")

	connectionSleep := ctx.Duration("sleep")

	benchmarker := NewWebsocketBenchmarker(
		WebsocketBenchmarkerOptionEndpoint(endpoint),
		WebsocketBenchmarkerOptionPath(path),
		WebsocketBenchmarkerOptionUserNum(userNum),
		WebsocketBenchmarkerOptionConnectInterval(connectInterval),
		WebsocketBenchmarkerConnectionSleep(connectionSleep),
	)

	log.Printf(benchmarker.ConnInfo())

	c := make(chan os.Signal, 1)
	signal.Notify(c, []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT}...)

	errC := make(chan error, 1)
	go func() {
		if err := benchmarker.Test(); err != nil {
			errC <- err
		}
		if err := benchmarker.StartConnBenchmark(); err != nil {
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
