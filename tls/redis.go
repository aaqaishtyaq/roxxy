package tls

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/go-redis/redis/v8"
	lru "github.com/hashicorp/golang-lru"
)

type RedisCertificateLoader struct {
	*redis.Client
	cache *lru.Cache
}

func NewRedisCertificateLoader(client *redis.Client) *RedisCertificateLoader {
	cache, _ := lru.New(100)
	return &RedisCertificateLoader{
		Client: client,
		cache:  cache,
	}
}

func (r *RedisCertificateLoader) GetCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert, err := r.getCertificateByCNAME(getWildCard(clientHello.ServerName))

	if _, ok := err.(ErrCertificateNotFound); ok {
		return r.getCertificateByCNAME(clientHello.ServerName)
	}

	return cert, err
}

func (r *RedisCertificateLoader) getCertificateByCNAME(serverName string) (*tls.Certificate, error) {
	if data, ok := r.cache.Get(serverName); ok {
		c := data.(certCached)
		if !c.Expired() {
			return c.cert, nil
		}
	}
	return r.getCertificateFromRedis(serverName)
}

func (r *RedisCertificateLoader) getCertificateFromRedis(serverName string) (*tls.Certificate, error) {
	ctx := context.Background()
	data, err := r.Client.HMGet(ctx, "tls:"+serverName, "certificate", "key").Result()
	if err != nil {
		return nil, err
	}

	var certificate string
	var key string

	if data[0] != nil {
		certificate = data[0].(string)
	}
	if data[1] != nil {
		key = data[1].(string)
	}

	if certificate == "" || key == "" {
		return nil, ErrCertificateNotFound{serverName}
	}
	cert, err := tls.X509KeyPair([]byte(certificate), []byte(key))
	if err != nil {
		return nil, err
	}
	r.cache.Add(serverName, certCached{
		cert:    &cert,
		expires: time.Now().Add(30 * time.Second),
	})
	return &cert, nil
}

type certCached struct {
	cert    *tls.Certificate
	expires time.Time
}

func (c *certCached) Expired() bool {
	return time.Now().After(c.expires)
}
