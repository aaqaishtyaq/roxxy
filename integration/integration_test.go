package integration

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aaqaishtyaq/roxxy/backend"
	"github.com/aaqaishtyaq/roxxy/reverseproxy"
	"github.com/aaqaishtyaq/roxxy/router"
	"github.com/go-redis/redis/v8"
)

const redisDB = 5

func clearKeys(r *redis.Client) error {
	ctx := context.Background()
	val := r.Keys(ctx, "frontend:*").Val()
	val = append(val, r.Keys(ctx, "dead:*").Val()...)
	if len(val) > 0 {
		return r.Del(ctx, val...).Err()
	}
	return nil
}

func redisConn() (*redis.Client, error) {
	return redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: redisDB}), nil
}

func initTest(t testing.TB) *redis.Client {
	redis, err := redisConn()
	if err != nil {
		t.Fatal(err)
	}
	err = clearKeys(redis)
	if err != nil {
		t.Fatal(err)
	}
	return redis
}

func BenchmarkFullStackNativeRedisNoCache(b *testing.B) {
	b.StopTimer()
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	r := initTest(b)
	backends := make([]interface{}, 100)
	defer srv.Close()
	for i := range backends {
		backends[i] = srv.URL
	}
	backends = append([]interface{}{"benchfrontend"}, backends...)
	ctx := context.Background()
	err := r.RPush(ctx, "frontend:myfrontend.com", backends...).Err()
	if err != nil {
		b.Fatal(err)
	}
	rp := &reverseproxy.NativeReverseProxy{}
	opts := backend.RedisOptions{
		DB: redisDB,
	}
	be, err := backend.NewRedisBackend(ctx, opts, opts)
	if err != nil {
		b.Fatal(err)
	}
	router := router.Router{
		Backend: be,
	}
	err = router.Init(ctx)
	if err != nil {
		b.Fatal(err)
	}
	err = rp.Initialize(reverseproxy.ReverseProxyConfig{
		Router: &router,
	})
	if err != nil {
		b.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	url := fmt.Sprintf("http://%s/", listener.Addr().String())
	go rp.Listen(listener, nil)
	defer rp.Stop()
	cli := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 1000,
		},
	}
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			request, _ := http.NewRequest("GET", url, nil)
			request.Host = "myfrontend.com"
			rsp, err := cli.Do(request)
			if rsp == nil || rsp.StatusCode != http.StatusOK {
				b.Fatalf("invalid response %#v: %s", rsp, err)
			}
			io.Copy(ioutil.Discard, rsp.Body)
			rsp.Body.Close()
		}
	})
	b.StopTimer()
}
