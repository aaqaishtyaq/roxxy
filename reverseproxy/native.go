package reverseproxy

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	noopDirector = func(*http.Request) {}

	openConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "roxxy",
		Subsystem: "reverseproxy",
		Name:      "connections_current_open",
		Help:      "The current number of open connections excluding hijacked ones.",
	})
)

type NativeReverseProxy struct {
	http.Transport
	ReverseProxyConfig
	servers []*http.Server
	rp      *httputil.ReverseProxy
	dialer  *net.Dialer
}

type bufferPool struct {
	syncPool sync.Pool
}

func (p *bufferPool) Get() []byte {
	b := p.syncPool.Get()
	if b == nil {
		return make([]byte, 32*1024)
	}
	return b.([]byte)
}

func (p *bufferPool) Put(b []byte) {
	p.syncPool.Put(b)
}

func (rp *NativeReverseProxy) Initialize(rpConfig ReverseProxyConfig) error {
	rp.ReverseProxyConfig = rpConfig
	rp.servers = make([]*http.Server, 0)

	rp.dialer = &net.Dialer{
		Timeout:   rp.DialTimeout,
		KeepAlive: 30 * time.Second,
	}

	rp.Transport = http.Transport{
		Dial:                rp.dialer.Dial,
		TLSHandshakeTimeout: rp.DialTimeout,
		MaxIdleConnsPerHost: 50,
		DisableCompression:  true,
	}

	rp.rp = &httputil.ReverseProxy{
		Director:      noopDirector,
		Transport:     rp,
		FlushInterval: rp.FlushInterval,
		BufferPool:    &bufferPool{},
	}

	return nil
}

func (rp *NativeReverseProxy) Listen(listener net.Listener, tlsConfig *tls.Config) {
	server := &http.Server{
		ReadTimeout:       rp.ReadTimeout,
		ReadHeaderTimeout: rp.ReadHeaderTimeout,
		WriteTimeout:      rp.WriteTimeout,
		IdleTimeout:       rp.IdleTimeout,
		Handler:           rp,
		TLSConfig:         tlsConfig,
		ConnState: func(c net.Conn, s http.ConnState) {
			switch s {
			case http.StateNew:
				openConnections.Inc()
			case http.StateHijacked:
				openConnections.Dec()
			case http.StateClosed:
				openConnections.Dec()
			}
		},
	}
	rp.servers = append(rp.servers, server)
	server.Serve(listener)
}

func (rp *NativeReverseProxy) Stop() {
	for _, server := range rp.servers {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		server.Shutdown(ctx)
		cancel()
	}
}

func (rp *NativeReverseProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Host == "__ping__" && req.URL.Path == "/" {
		err := rp.Router.HealthCheck()
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			return
		}
		rw.WriteHeader(http.StatusOK)
		rw.Write(okResponse)
		return
	}
	if rp.RequestIDHeader != "" && headerGet(req.Header, rp.RequestIDHeader) == "" {
		unparsedID := uuid.New()
		headerSet(req.Header, rp.RequestIDHeader, unparsedID.String())
	}

	upgrade
}

func headerGet(header http.Header, key string) string {
	if header == nil {
		return ""
	}
	entry := header[key]
	if len(entry) == 0 {
		return ""
	}
	return entry[0]
}

func headerSet(header http.Header, key, value string) {
	header[key] = []string{value}
}
