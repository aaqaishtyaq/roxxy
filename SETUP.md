# Setup

```bash
# Run redis
% docker run -d -p 6379:6379 redis

# Run roxxy
% ./roxxy --read-redis-host 127.0.0.1 --write-redis-host 127.0.0.1 --listen ":80"

## Add values to redis
% redis-cli
127.0.0.1:6379> rpush frontend:demo.roxxy.sh demoInstance2
127.0.0.1:6379> rpush frontend:demo.roxxy.sh http://127.0.0.1:8000
127.0.0.1:6379> lrange frontend:demo.roxxy.sh 0 -1
1) "demoInstance2"
2) "http://127.0.0.1:8000"

# Run nginx
% docker run --rm --name nginx-serv -p 8000:80 -d nginx

# update /etc/hosts to points demo.roxxy.sh to localhost

...
demo.roxxy.dev    12.0.0.0.1
```
