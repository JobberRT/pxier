package core

import (
	"crypto/tls"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"strings"
	"time"
)

type cplFetcher struct {
	url     string
	timeout time.Duration
	client  *fasthttp.Client
}

func newCPLFetcher(url, proxy string, timeO int64) *cplFetcher {
	if len(url) == 0 {
		url = "https://raw.githubusercontent.com/clarketm/proxy-list/master/proxy-list-raw.txt"
	}
	if timeO == 0 {
		timeO = 5
	}
	f := &cplFetcher{
		url:     url,
		timeout: time.Duration(timeO) * time.Second,
		client:  &fasthttp.Client{},
	}
	if len(proxy) != 0 {
		if strings.Contains(proxy, "http") {
			f.client = &fasthttp.Client{
				Dial: fasthttpproxy.FasthttpHTTPDialer(proxy),
			}
		} else {
			f.client = &fasthttp.Client{
				Dial: fasthttpproxy.FasthttpSocksDialer(proxy),
			}
		}
	} else {
		f.client = &fasthttp.Client{
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	return f
}

func (f *cplFetcher) Fetch() []*Proxy {
	logrus.WithField("provider", f.Type()).Info("fetch")
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(f.url)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.SetContentEncoding("gzip")
	if err := f.client.DoTimeout(req, res, f.timeout); err != nil {
		logrus.WithError(err).WithField("url", f.url).Error("failed to get proxy")
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
			Provider:  ProviderTypeCPL,
			DialType:  ProxyTypeHttp,
		})
	}
	return proxies
}

func (f *cplFetcher) Type() string {
	return ProviderTypeCPL
}
