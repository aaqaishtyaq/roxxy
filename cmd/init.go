package cmd

import (
	"log"
	"os"

	"github.com/aaqaishtyaq/roxxy/reverseproxy"
	"github.com/google/gops/agent"
	"github.com/urfave/cli/v2"
)

func runServer(c *cli.Context) error {
	err := agent.Listen(agent.Options{})
	if err != nil {
		log.Printf("Unable to start gops agent: %v", err)
	}

	var rp reverseproxy.ReverseProxy
	rp = &reverseproxy.NativeReverseProxy{}

	return nil
}

func Execute() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "listen",
			Aliases: []string{"l"},
			Value:   "0.0.0.0:8989",
			Usage:   "Address to listen",
		},
		&cli.StringFlag{
			Name:  "read-redis-network",
			Value: "tcp",
			Usage: "Redis address network, possible values are \"tcp\" for tcp connection and \"unix\" for connecting using unix sockets",
		},
		&cli.StringFlag{
			Name:  "read-redis-host",
			Value: "127.0.0.1",
			Usage: "Redis host address for tcp connections or socket path for unix sockets",
		},
		&cli.IntFlag{
			Name:  "read-redis-port",
			Value: 6379,
			Usage: "Redis port",
		},
		&cli.StringFlag{
			Name:  "read-redis-sentinel-addrs",
			Usage: "Comma separated list of redis sentinel addresses",
		},
		&cli.StringFlag{
			Name:  "read-redis-sentinel-name",
			Usage: "Redis sentinel name",
		},
		&cli.StringFlag{
			Name:  "read-redis-password",
			Usage: "Redis password",
		},
		&cli.IntFlag{
			Name:  "read-redis-db",
			Usage: "Redis database number",
		},
		&cli.StringFlag{
			Name:  "write-redis-network",
			Value: "tcp",
			Usage: "Redis address network, possible values are \"tcp\" for tcp connection and \"unix\" for connecting using unix sockets",
		},
		&cli.StringFlag{
			Name:  "write-redis-host",
			Value: "127.0.0.1",
			Usage: "Redis host address for tcp connections or socket path for unix sockets",
		},
		&cli.IntFlag{
			Name:  "write-redis-port",
			Value: 6379,
			Usage: "Redis port",
		},
		&cli.StringFlag{
			Name:  "write-redis-sentinel-addrs",
			Usage: "Comma separated list of redis sentinel addresses",
		},
		&cli.StringFlag{
			Name:  "write-redis-sentinel-name",
			Usage: "Redis sentinel name",
		},
		&cli.StringFlag{
			Name:  "write-redis-password",
			Usage: "Redis password",
		},
		&cli.IntFlag{
			Name:  "write-redis-db",
			Usage: "Redis database number",
		},
		&cli.IntFlag{
			Name:  "request-timeout",
			Value: 30,
			Usage: "Total backend request timeout in seconds",
		},
		&cli.IntFlag{
			Name:  "dial-timeout",
			Value: 10,
			Usage: "Dial backend request timeout in seconds",
		},
		&cli.DurationFlag{
			Name:  "client-read-timeout",
			Value: 0,
			Usage: "Maximum duration for reading the entire request, including the body",
		},
		&cli.DurationFlag{
			Name:  "client-read-header-timeout",
			Value: 0,
			Usage: "Amount of time allowed to read request headers",
		},
		&cli.DurationFlag{
			Name:  "client-write-timeout",
			Value: 0,
			Usage: "Maximum duration before timing out writes of the response",
		},
		&cli.DurationFlag{
			Name:  "client-idle-timeout",
			Value: 0,
			Usage: "Maximum amount of time to wait for the next request when keep-alives are enabled",
		},
		&cli.IntFlag{
			Name:  "dead-backend-time",
			Value: 30,
			Usage: "Time in seconds a backend will remain disabled after a network failure",
		},
		&cli.IntFlag{
			Name:  "flush-interval",
			Value: 10,
			Usage: "Time in milliseconds to flush the proxied request",
		},
		&cli.StringFlag{
			Name:  "request-id-header",
			Usage: "Header to enable message tracking",
		},
		&cli.BoolFlag{
			Name:  "active-healthcheck",
			Usage: "Enable active healthcheck on dead backends once they are marked as dead. Enabling this flag will result in dead backends only being enabled again once the active healthcheck routine is able to reach them.",
		},
		&cli.BoolFlag{
			Name:  "backend-cache",
			Usage: "Enable caching backend results for 2 seconds. This may cause temporary inconsistencies.",
		},
	}
	app.Name = "roxxy"
	app.Usage = "http and websockets reverse proxy"
	app.Version = "0.0.1"
	app.Action = runServer

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
