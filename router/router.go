package router

import (
	"sync"
	"time"

	"github.com/aaqaishtyaq/roxxy/backend"
	"github.com/aaqaishtyaq/roxxy/log"
	lru "github.com/hashicorp/golang-lru"
)

var (
	cacheTTLExpires = 2 * time.Second
)

type Router struct {
	LogPath        string
	DeadBackendTTL int
	Backend        backend.RoutesBackend
	CacheEnabled   bool
	logger         *log.Logger
	rrMutex        sync.RWMutex
	roundRobin     map[string]*uint32
	cache          *lru.Cache
}

type backendSet struct {
	id       string
	backends []string
	dead     map[int]struct{}
	expires  time.Time
}

func (s *backendSet) Expired() bool {
	return time.Now().After(s.expires)
}

func (router *Router) Init() error {
	var err error

	if router.Backend == nil {
		var be backend.RoutesBackend
		be, err = backend.NewRedisBackend(backend.RedisOptions{}, backend.RedisOptions{})
		if err != nil {
			return err
		}
		router.Backend = be
	}

	if router.LogPath == "" {
		router.LogPath = "none"
	}

	if router.logger == nil && router.LogPath != "none" {
		router.logger, err = log.NewFileLogger(router.LogPath)
		if err != nil {
			return err
		}
	}

	if router.DeadBackendTTL == 0 {
		router.DeadBackendTTL = 30
	}

	if router.CacheEnabled && router.cache == nil {
		router.cache, err = lru.New(100)
		if err != nil {
			return err
		}
	}

	router.roundRobin = make(map[string]*uint32)
	return nil
}
