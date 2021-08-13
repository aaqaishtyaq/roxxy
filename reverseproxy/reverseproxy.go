package reverseproxy

import (
	"crypto/tls"
	"net"
	"time"
)

type Router interface {
	HealthCheck() error
	ChooseBackend(host string) (*RequestData, error)
	EndRequest(reqData *RequestData, isDead bool) error
}

type ReverseProxy interface {
	Initialize(rpConfig ReverseProxyConfig) error
	Listen(net.Listener, *tls.Config)
	Stop()
}

type RequestData struct {
	BackendLen int
	Backend    string
	BackendIdx int
	BackendKey string
	Host       string
	StartTime  time.Time
	AllDead    bool
}

type ReverseProxyConfig struct {
	Router            Router
	FlushInterval     time.Duration
	DialTimeout       time.Duration
	RequestTimeout    time.Duration
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	RequestIDHeader   string
}
