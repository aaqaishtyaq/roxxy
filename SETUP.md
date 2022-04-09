# Setup

```bash
# Run redis
% docker run -d -p 6379:6379 redis

# Run roxxy
% ./roxxy --read-redis-host 127.0.0.1 --write-redis-host 127.0.0.1 --listen ":80"

## Add values to redis
% redis-cli
127.0.0.1:6379> rpush frontend:demo.aaqa.dev demoInstance2
127.0.0.1:6379> rpush frontend:demo.aaqa.dev http://127.0.0.1:8000
127.0.0.1:6379> lrange frontend:demo.aaqa.dev 0 -1
1) "demoInstance2"
2) "http://127.0.0.1:8000"

## Run nginx
% docker run --rm --name nginx-serv -p 8000:80 -d nginx
```

<!-- docker run --rm -d --net=host roxxy --listen ":80" --listen ":80" --read-redis-host 172.17.0.2 --write-redis-host 172.17.0.2 -->
