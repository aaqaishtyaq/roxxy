package backend

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type redisBackend struct {
	readClient  *redis.Client
	writeClient *redis.Client
	monitor     *redisMonitor
}

type RedisOptions struct {
	Network       string
	Host          string
	Port          int
	SentinelAddrs string
	SentinelName  string
	Password      string
	DB            int
}

const (
	dialTimeout  = time.Second
	readTimeout  = time.Second
	writeTimeout = time.Second
	poolTimeout  = time.Second
	poolSize     = 1000
	idleTimeout  = time.Minute
	maxRetries   = 1
)

func (opts RedisOptions) Client() (*redis.Client, error) {
	if opts.SentinelAddrs == "" {
		if opts.Host == "" {
			opts.Host = "127.0.0.1"
		}
		if opts.Port == 0 {
			opts.Port = 6379
		}
		var addr string
		if opts.Network == "unix" {
			addr = opts.Host
		} else {
			addr = fmt.Sprintf("%s:%d", opts.Host, opts.Port)
		}
		return redis.NewClient(&redis.Options{
			Network:      opts.Network,
			Addr:         addr,
			Password:     opts.Password,
			DB:           opts.DB,
			MaxRetries:   maxRetries,
			DialTimeout:  dialTimeout,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			PoolSize:     poolSize,
			PoolTimeout:  poolTimeout,
			IdleTimeout:  idleTimeout,
		}), nil
	}
	addresses := strings.Split(opts.SentinelAddrs, ",")
	for i := range addresses {
		addresses[i] = strings.TrimSpace(addresses[i])
		if addresses[i] == "" {
			return nil, errors.New("redis sentinel address connot be empty")
		}
	}
	if opts.SentinelName == "" {
		return nil, errors.New("redis sentinel name cannot be empty")
	}
	return redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    opts.SentinelName,
		SentinelAddrs: addresses,
		Password:      opts.Password,
		DB:            opts.DB,
		MaxRetries:    maxRetries,
		DialTimeout:   dialTimeout,
		ReadTimeout:   readTimeout,
		WriteTimeout:  writeTimeout,
		PoolSize:      poolSize,
		PoolTimeout:   poolTimeout,
		IdleTimeout:   idleTimeout,
	}), nil
}

func NewRedisBackend(ctx context.Context, readOpts, writeOpts RedisOptions) (RoutesBackend, error) {
	rClient, err := readOpts.Client()
	if err != nil {
		return nil, err
	}
	err = rClient.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}
	wClient, err := writeOpts.Client()
	if err != nil {
		return nil, err
	}
	err = wClient.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}
	return &redisBackend{
		readClient:  rClient,
		writeClient: wClient,
	}, nil
}

func (b *redisBackend) Healthcheck(ctx context.Context) error {
	return b.readClient.Ping(ctx).Err()
}

func (b *redisBackend) Backends(ctx context.Context, host string) (string, []string, map[int]struct{}, error) {
	pipe := b.readClient.Pipeline()
	defer pipe.Close()
	rangeVal := pipe.LRange(ctx, "frontend:"+host, 0, -1)
	membersVal := pipe.SMembers(ctx, "dead:"+host)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return "", nil, nil, err
	}
	deadMap := map[int]struct{}{}
	for _, item := range membersVal.Val() {
		intVal, _ := strconv.ParseInt(item, 10, 32)
		deadMap[int(intVal)] = struct{}{}
	}
	backends := rangeVal.Val()
	if len(backends) < 2 {
		return "", nil, nil, ErrNoBackends
	}
	return host, backends[1:], deadMap, nil
}

func (b *redisBackend) MarkDead(ctx context.Context, host string, backend string, backendIdx int, backendLen int, deadTTL int) error {
	pipe := b.writeClient.Pipeline()
	defer pipe.Close()
	deadKey := "dead:" + host
	pipe.SAdd(ctx, deadKey, backendIdx)
	pipe.Expire(ctx, deadKey, time.Duration(deadTTL)*time.Second)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	deadMsg := fmt.Sprintf("%s;%s;%d;%d", host, backend, backendIdx, backendLen)
	return b.writeClient.Publish(ctx, "dead", deadMsg).Err()
}

func (b *redisBackend) StartMonitor(ctx context.Context) error {
	var err error
	b.monitor, err = newRedisMonitor(ctx, b.writeClient)
	return err
}

func (b *redisBackend) StopMonitor() {
	if b.monitor != nil {
		b.monitor.stop()
	}
}
