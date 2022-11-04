package main

import (
	"github.com/urfave/cli/v2"
	"os"
)

var GlobalFlags = []cli.Flag{
	&cli.StringFlag{Name: "host", Value: "localhost:8080"},
	&cli.StringFlag{Name: "path", Value: "/ws"},
	&cli.IntFlag{Name: "user", Value: 500},
	&cli.StringFlag{Name: "connectInterval", Value: "20ms"},
}

func main() {

	app := cli.NewApp()

	app.Name = "websocket-benchmark-cli"
	app.Usage = "websocket benchmark tool"
	app.Version = "v1.0.0"
	app.Flags = GlobalFlags
	app.Commands = []*cli.Command{
		CmdConn,
		CmdMessage,
	}
	app.AllowExtFlags = true
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
