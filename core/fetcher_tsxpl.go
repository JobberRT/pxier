package core

import (
	"crypto/tls"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"strings"
	"time"
)

type tsxFetcher struct {
	httpUrl   string
	socks5Url string
	client    *fasthttp.Client
	timeout   time.Duration
}

func newTSXFetcher(hu, su, proxy string, timeO int64) *tsxFetcher {
	if len(hu) == 0 {
		hu = "https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/http.txt"
	}
	if len(su) == 0 {
		su = "https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/socks5.txt"
	}

	if timeO == 0 {
		timeO = 5
	}
	f := &tsxFetcher{
		httpUrl:   hu,
		socks5Url: su,
		timeout:   time.Duration(timeO) * time.Second,
		client:    &fasthttp.Client{TLSConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	if len(proxy) != 0 {
		if strings.Contains(proxy, "http") {
			f.client.Dial = fasthttpproxy.FasthttpHTTPDialer(proxy)
		} else {
			f.client.Dial = fasthttpproxy.FasthttpSocksDialer(proxy)
		}
	}
	return f
}

func (f *tsxFetcher) Fetch() []*Proxy {
	logrus.WithField("provider", f.Type()).Info("fetch")
	return append(f.fetchHttpProxy(), f.fetchSocks5Proxy()...)
}

func (f *tsxFetcher) Type() string {
	return ProviderTypeTSXPL
}

func (f *tsxFetcher) fetchHttpProxy() []*Proxy {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(f.httpUrl)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.SetContentEncoding("gzip")
	if err := f.client.DoTimeout(req, res, f.timeout); err != nil {
		logrus.WithError(err).WithField("url", f.httpUrl).Error("failed to get proxy")
	}

	body, err := readBody(res)
	if err != nil {
		logrus.WithError(err).WithField("raw", string(res.Body())).Error("failed to unGzip body")
		return nil
	}
	rawSlice := strings.Split(string(body), "\n")
	proxies := make([]*Proxy, 0)
	for _, each := range rawSlice {
		if len(each) == 0 {
			continue
		}
		proxies = append(proxies, &Proxy{
			Address:   each,
			ErrTimes:  0,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			Provider:  ProviderTypeTSXPL,
			DialType:  ProxyTypeHttp,
		})
	}
	return proxies
}

func (f *tsxFetcher) fetchSocks5Proxy() []*Proxy {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(f.socks5Url)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.SetContentEncoding("gzip")
	if err := f.client.DoTimeout(req, res, f.timeout); err != nil {
		logrus.WithError(err).WithField("url", f.socks5Url).Error("failed to get proxy")
	}

	body, err := readBody(res)
	if err != nil {
		logrus.WithError(err).WithField("raw", string(res.Body())).Error("failed to unGzip body")
		return nil
	}
	rawSlice := strings.Split(string(body), "\n")
	proxies := make([]*Proxy, 0)
	for _, each := range rawSlice {
		if len(each) == 0 {
			continue
		}
		proxies = append(proxies, &Proxy{
			Address:   each,
			ErrTimes:  0,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			Provider:  ProviderTypeTSXPL,
			DialType:  ProxyTypeSocks5,
		})
	}
	return proxies
}
