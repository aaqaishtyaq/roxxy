package backend

import "errors"

var ErrNoBackends = errors.New("no backends")

type RoutesBackend interface {
	Healthcheck() error
	Backends(host string) (string, []string, map[int]struct{}, error)
	MarkDead(host string, backend string, backendIdx int, backendLen int, deadTTL int) error
	StartMonitor() error
	StopMonitor()
}
