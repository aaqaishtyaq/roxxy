package backend

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"gopkg.in/check.v1"
)

type S struct {
	redisConn *redis.Client
	be        RoutesBackend
}

var _ = check.Suite(&S{})

func Test(t *testing.T) {
	check.TestingT(t)
}

func (s *S) SetUpTest(c *check.C) {
	ctx := context.Background()
	s.redisConn = redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 1})
	val := s.redisConn.Keys(ctx, "frontend:*").Val()
	val = append(val, s.redisConn.Keys(ctx, "dead:*").Val()...)
	var err error
	if len(val) > 0 {
		err = s.redisConn.Del(ctx, val...).Err()
		c.Assert(err, check.IsNil)
	}
	s.be, err = NewRedisBackend(ctx, RedisOptions{DB: 1}, RedisOptions{DB: 1})
	c.Assert(err, check.IsNil)
}

func (s *S) TearDownTest(c *check.C) {
	s.redisConn.Close()
}

func (s *S) TestBackends(c *check.C) {
	ctx := context.Background()
	err := s.redisConn.RPush(ctx, "frontend:f1.com", "f1.com", "srv1", "srv2").Err()
	c.Assert(err, check.IsNil)
	key, backends, deadMap, err := s.be.Backends(ctx, "f1.com")
	c.Assert(err, check.IsNil)
	c.Assert(key, check.Equals, "f1.com")
	c.Assert(backends, check.DeepEquals, []string{"srv1", "srv2"})
	c.Assert(deadMap, check.DeepEquals, map[int]struct{}{})
}

func (s *S) TestBackendsIgnoresName(c *check.C) {
	ctx := context.Background()
	err := s.redisConn.RPush(ctx, "frontend:f1.com", "xxxxxxx", "srv1", "srv2").Err()
	c.Assert(err, check.IsNil)
	key, backends, deadMap, err := s.be.Backends(ctx, "f1.com")
	c.Assert(err, check.IsNil)
	c.Assert(key, check.Equals, "f1.com")
	c.Assert(backends, check.DeepEquals, []string{"srv1", "srv2"})
	c.Assert(deadMap, check.DeepEquals, map[int]struct{}{})
}

func (s *S) TestBackendsWithDead(c *check.C) {
	ctx := context.Background()
	err := s.redisConn.RPush(ctx, "frontend:f1.com", "xxxxxxx", "srv1", "srv2").Err()
	c.Assert(err, check.IsNil)
	err = s.be.MarkDead(ctx, "f1.com", "srv1", 0, 2, 30)
	c.Assert(err, check.IsNil)
	err = s.be.MarkDead(ctx, "f1.com", "srv2", 1, 2, 30)
	c.Assert(err, check.IsNil)
	key, backends, deadMap, err := s.be.Backends(ctx, "f1.com")
	c.Assert(err, check.IsNil)
	c.Assert(key, check.Equals, "f1.com")
	c.Assert(backends, check.DeepEquals, []string{"srv1", "srv2"})
	c.Assert(deadMap, check.DeepEquals, map[int]struct{}{0: {}, 1: {}})
}

func (s *S) TestMarkDead(c *check.C) {
	ctx := context.Background()
	pubsub := s.redisConn.Subscribe(ctx, "dead")
	err := s.be.MarkDead(ctx, "f1.com", "url1", 0, 2, 30)
	c.Assert(err, check.IsNil)
	members, err := s.redisConn.SMembers(ctx, "dead:f1.com").Result()
	c.Assert(err, check.IsNil)
	c.Assert(members, check.DeepEquals, []string{"0"})
	msg, err := pubsub.ReceiveMessage(ctx)
	c.Assert(err, check.IsNil)
	c.Assert(msg.Payload, check.Equals, "f1.com;url1;0;2")
}
