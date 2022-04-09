package reverseproxy

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"time"

	"github.com/aaqaishtyaq/roxxy/log"
)

var (
	noRouteResponseContent = []byte("no such route")
	allBackendsDeadContent = []byte("all backends are dead")
	okResponse             = []byte("OK")

	ErrAllBackendsDead      = errors.New(string(allBackendsDeadContent))
	ErrNoRegisteredBackends = errors.New("no backends registered for host")
)

type Router interface {
	Healthcheck(ctx context.Context) error
	ChooseBackend(ctx context.Context, host string) (*RequestData, error)
	EndRequest(ctx context.Context, reqData *RequestData, isDead bool, fn func() *log.LogEntry) error
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
