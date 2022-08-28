package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type TargetHost struct {
	Host        string
	IsHttps     bool
	CAPath      string
	ConnTimeOut time.Duration
	ReadTimeOut time.Duration
}

func relay(c *gin.Context, targetHost *TargetHost) {
	if targetHost.ConnTimeOut == 0 {
		targetHost.ConnTimeOut = 5 * time.Second
	}
	if targetHost.ReadTimeOut == 0 {
		targetHost.ReadTimeOut = 10 * time.Second
	}
	hostReverseProxy(c, targetHost)
}

func hostReverseProxy(c *gin.Context, targetHost *TargetHost) {
	scheme := ""
	if targetHost.IsHttps {
		scheme += "https://"
	} else {
		scheme += "http://"
	}
	remote, err := url.Parse(scheme + targetHost.Host)
	if err != nil {
		logrus.Info("relay error, err(%v)", err)
		c.Status(http.StatusBadGateway)
		c.Abort()
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	if targetHost.IsHttps {
		tlsConfig, err := getVerTLSConfig(targetHost.CAPath)
		if err != nil {
			logrus.Info("relay error, err(%v)", err)
			c.Status(http.StatusBadGateway)
			c.Abort()
			return
		}
		pTransport := &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(netw, addr, targetHost.ConnTimeOut);
				if err != nil {
					return nil, err
				}
				return c, nil
			},
			ResponseHeaderTimeout: targetHost.ReadTimeOut,
			TLSClientConfig:       tlsConfig,
		}
		proxy.Transport = pTransport
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

func getVerTLSConfig(CAPath string) (*tls.Config, error) {
	caData, err := ioutil.ReadFile(CAPath)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caData)

	tlsConfig := &tls.Config{
		RootCAs: pool,
	}
	return tlsConfig, nil
}
