# Roxxy: A distributed HTTP and websockets reverse proxy

Roxxy is a distributed HTTP/websockets reverse proxy backed by Redis inspired by [planb](https://github.com/tsuru/planb) and [hipache](https://travis-ci.org/tsuru/planb)

## Features (TODO)

- [x] Load-Balancing

- [x] Dead Backend Detection

- [x] Dynamic Configuration

- [x] WebSocket

- [x] TLS

## Install

The easiest way to install `Roxxy Proxy` is to pull the build from the `Github Container Registry` and launch it in the container:

```console
# run Redis

docker run -d -p 6379:6379 redis

# run Roxxy
docker run -d --net=host ghcr.io/aaqaishtyaq/roxxy --listen ":80"
```

## VHOST Configuration

The configuration is managed by **Redis** that makes possible
to update the configuration dynamically and gracefully while
the server is running, and have that state shared across workers
and even across instances.

Let's take an example to proxify requests to 2 backends for the hostname
`www.aaqa.dev`. The 2 backends IP are `10.10.0.2` and `10.10.0.3` and
they serve the HTTP traffic on the port `80`.

`redis-cli` is the standard client tool to talk to Redis from the terminal.

Follow these steps:

### Create the frontend

```console
$ redis-cli rpush frontend:www.aaqa.dev mywebsite
(integer) 1
```

The frontend identifer is `mywebsite`, it could be anything.

### Add the 2 backends

```console
$ redis-cli rpush frontend:www.aaqa.dev http://10.10.0.2:80
(integer) 2
$ redis-cli rpush frontend:www.aaqa.dev http://10.10.0.3:80
(integer) 3
```

### Review the configuration

```console
$ redis-cli lrange frontend:www.aaqa.dev 0 -1
1) "mywebsite"
2) "http://10.10.0.2:80"
3) "http://10.10.0.3:80"
```

### TLS Configuration using redis (optional)

```console
redis-cli -x hmset tls:www.aaqa.dev certificate < server.crt
redis-cli -x hmset tls:www.aaqa.dev key < server.key

redis-cli -x hmset tls:*.tsuru.com certificate < wildcard.crt
redis-cli -x hmset tls:*.tsuru.io key < wildcard.key
```

### TLS Configuration using FS (optional)

create directory following this structure

```console
cd certficates
ls
*.domain-wildcard.com.key
*.domain-wildcard.com.crt
absolute-domain.key
absolute-domain.crt
```

While the server is running, any of these steps can be
re-run without messing up with the traffic.


## Start-up flags

The following flags are available for configuring Roxxy on start-up:

| **Flag**  | **Description**  |
|--- |--- |
| `--listen value, -l value`  | Address to listen.<br><br>(default: "0.0.0.0:8989")  |
| `-tls-listen value`  | Address to listen with tls.  |
| `--tls-preset value`  | Preset containing supported TLS versions and cyphers, according <br>to <https://wiki.mozilla.org/Security/Server_Side_TLS>. Possible  |
| `--metrics-address value`  | Address to expose Prometheus.<br><br>(default: "/metrics")  |
| `--load-certificates-from value`  | Path where certificate will found. If value equals 'redis'<br>certificate will be loaded from redis service. <br><br>(default: "redis")  |
| `--read-redis-network value`  | Redis address network, possible values are "tcp" for TCP<br>connection and "unix" for connecting using unix sockets.<br><br>(default: "tcp")  |
| `--read-redis-host value`  | Redis host address for tcp connections or socket path <br>for UNIX sockets. <br><br>(default: "127.0.0.1")  |
| `--read-redis-port value`  | Redis port<br><br>(default: 6379)  |
| `--read-redis-sentinel-addrs value`  | Comma separated list of redis sentinel addresses.  |
| `--read-redis-sentinel-name value`  | Redis sentinel name  |
| `--read-redis-password value`  | Redis password  |
| `--read-redis-db value`  | Redis database number <br><br>(default: 0)  |
| `--write-redis-network value`  | Redis address network, possible values are "tcp" for TCP<br>connection and "unix" for connecting using unix sockets<br><br>(default: "tcp")  |
| `--write-redis-host value`  | Redis host address for tcp connections or socket path <br>for UNIX sockets <br><br>(default: "127.0.0.1")  |
| `--write-redis-port value`  | Redis port <br><br>(default: 6379)  |
| `--write-redis-sentinel-addrs value`  | Comma separated list of redis sentinel addresses  |
| `--write-redis-sentinel-name value`  | Redis sentinel name  |
| `--write-redis-password value`  | Redis password  |
| `--write-redis-db value`  | Redis database number (default: 0)  |
| `--access-log value`  | File path where access log will be written. If value <br>equals 'syslog' log will be sent to local syslog. <br>The value 'none' can be used to disable access logs. <br><br>(default: "./access.log")  |
| `--request-timeout value`  | Total backend request timeout in seconds <br><br>(default: 30)  |
| `--dial-timeout value`  | Dial backend request timeout in seconds <br><br>(default: 10)  |
| `--client-read-timeout value`  | Maximum duration for reading the entire request, <br>including the body <br><br>(default: 0s)  |
| `--client-read-header-timeout value`  | Amount of time allowed to read request headers <br><br>(default: 0s)  |
| `--client-write-timeout value`  | Maximum duration before timing out writes of the response<br><br>(default: 0s)  |
| `--client-idle-timeout value`  | Maximum amount of time to wait for the next request <br>when keep-alives are enabled.<br><br>(default: 0s)  |
| `--dead-backend-time value`  | Time in seconds a backend will remain disabled after a <br>network failure. <br><br>(default: 30)  |
| `--flush-interval value`  | Time in milliseconds to flush the proxied request <br><br>(default: 10)  |
| `--request-id-header value`  | Header to enable message tracking  |
| `--active-healthcheck`  | Enable active healthcheck on dead backends once <br>they are marked as dead. Enabling this flag will<br>result in dead backends only being enabled again <br>once the active healthcheck routine is able to <br>reach them.  |
| `--backend-cache`  | Enable caching backend results for 2 seconds. <br>This may cause temporary inconsistencies.  |
| `--help, -h`  | show help  |
| `--version, -v`  | print the version  |
