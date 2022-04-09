package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/aaqaishtyaq/roxxy/backend"
	"github.com/aaqaishtyaq/roxxy/reverseproxy"
	"github.com/aaqaishtyaq/roxxy/router"
	"github.com/aaqaishtyaq/roxxy/tls"
	"github.com/google/gops/agent"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
)

func handleSignals(server interface {
	Stop()
},
) {
	sigChan := make(chan os.Signal, 3)
	go func() {
		for sig := range sigChan {
			if sig == os.Interrupt || sig == os.Kill {
				server.Stop()
				agent.Close()
			}
			if sig == syscall.SIGUSR1 {
				pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)
			}
			if sig == syscall.SIGUSR2 {
				go startProfiling()
			}
		}
	}()
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)
}

func startProfiling() {
	cpufile, _ := os.OpenFile("./roxxy_cpu.pprof", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o660)
	memfile, _ := os.OpenFile("./roxxy_mem.pprof", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o660)
	lockfile, _ := os.OpenFile("./roxxy_lock.pprof", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o660)
	log.Println("enabling profile...")
	runtime.GC()
	pprof.WriteHeapProfile(memfile)
	memfile.Close()
	runtime.SetBlockProfileRate(1)
	time.Sleep(30 * time.Second)
	pprof.Lookup("block").WriteTo(lockfile, 0)
	runtime.SetBlockProfileRate(0)
	lockfile.Close()
	pprof.StartCPUProfile(cpufile)
	time.Sleep(30 * time.Second)
	pprof.StopCPUProfile()
	cpufile.Close()
	log.Println("profiling done")
}

func runServer(c *cli.Context) error {
	err := agent.Listen(agent.Options{})
	if err != nil {
		log.Printf("Unable to start gops agent: %v", err)
	}

	rp := &reverseproxy.NativeReverseProxy{}

	readOpts := backend.RedisOptions{
		Network:       c.String("read-redis-network"),
		Host:          c.String("read-redis-host"),
		Port:          c.Int("read-redis-port"),
		SentinelAddrs: c.String("read-redis-sentinel-addrs"),
		SentinelName:  c.String("read-redis-sentinel-name"),
		Password:      c.String("read-redis-password"),
		DB:            c.Int("read-redis-db"),
	}

	writeOpts := backend.RedisOptions{
		Network:       c.String("write-redis-network"),
		Host:          c.String("write-redis-host"),
		Port:          c.Int("write-redis-port"),
		SentinelAddrs: c.String("write-redis-sentinel-addrs"),
		SentinelName:  c.String("write-redis-sentinel-name"),
		Password:      c.String("write-redis-password"),
		DB:            c.Int("write-redis-db"),
	}

	ctx := context.Background()

	routesBE, err := backend.NewRedisBackend(ctx, readOpts, writeOpts)
	if err != nil {
		log.Fatal(err)
	}

	if c.Bool("active-healthcheck") {
		err = routesBE.StartMonitor(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}

	r := router.Router{
		Backend:        routesBE,
		LogPath:        c.String("access-log"),
		DeadBackendTTL: c.Int("dead-backend-time"),
		CacheEnabled:   c.Bool("backend-cache"),
	}

	err = r.Init(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = rp.Initialize(reverseproxy.ReverseProxyConfig{
		Router:            &r,
		RequestIDHeader:   http.CanonicalHeaderKey(c.String("request-id-header")),
		FlushInterval:     time.Duration(c.Int("flush-interval")) * time.Millisecond,
		DialTimeout:       time.Duration(c.Int("dial-timeout")) * time.Second,
		RequestTimeout:    time.Duration(c.Int("request-timeout")) * time.Second,
		ReadTimeout:       c.Duration("client-read-timeout"),
		ReadHeaderTimeout: c.Duration("client-read-header-timeout"),
		WriteTimeout:      c.Duration("client-write-timeout"),
		IdleTimeout:       c.Duration("client-idle-timeout"),
	})

	if err != nil {
		log.Fatal(err)
	}

	listener := &router.RouterListener{
		ReverseProxy: rp,
		Listen:       c.String("listen"),
		TLSListen:    c.String("tls-listen"),
		TLSPreset:    c.String("tls-preset"),
		CertLoader:   getCertificateLoader(c, readOpts),
	}

	if addr := c.String("metrics-address"); addr != "" {
		handler := http.NewServeMux()
		handler.Handle("/metrics", promhttp.Handler())
		go func() {
			log.Fatal(http.ListenAndServe(addr, handler))
		}()
	}

	handleSignals(listener)
	listener.Serve()

	r.Stop()
	routesBE.StopMonitor()
	return nil
}

func getCertificateLoader(c *cli.Context, readOpts backend.RedisOptions) tls.CertificateLoader {
	if c.String("tls-listen") == "" {
		return nil
	}

	from := c.String("load-certificates-from")
	switch from {
	case "redis":
		client, err := readOpts.Client()
		if err != nil {
			log.Fatal(err)
		}

		return tls.NewRedisCertificateLoader(client)
	default:
		return tls.NewFSCertificateLoader(from)
	}
}

func Execute() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "listen",
			Aliases: []string{"l"},
			Value:   ":8989",
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
