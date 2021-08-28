package reverseproxy

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/aaqaishtyaq/roxxy/log"
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

func (r *RequestData) logError(path string, rid string, err error) {
	log.ErrorLogger.MessageRaw(&log.LogEntry{
		Err: &log.ErrEntry{
			Backend: r.Backend,
			Host:    r.Host,
			Path:    path,
			Rid:     rid,
			Err:     err.Error(),
		},
	})
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
