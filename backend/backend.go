package backend

import (
	"context"
	"errors"
)

var ErrNoBackends = errors.New("no backends")

type RoutesBackend interface {
	Healthcheck(ctx context.Context) error
	Backends(ctx context.Context, host string) (string, []string, map[int]struct{}, error)
	MarkDead(ctx context.Context, host string, backend string, backendIdx int, backendLen int, deadTTL int) error
	StartMonitor(ctx context.Context) error
	StopMonitor()
}
