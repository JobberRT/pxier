package core

import (
	"crypto/tls"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"regexp"
	"strings"
	"time"
)

var (
	keyPattern *regexp.Regexp
	ipPattern  *regexp.Regexp
)

func init() {
	keyPattern = regexp.MustCompile("[a-z\\d]{32}")
	ipPattern = regexp.MustCompile("\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}:\\d{1,5}")
}

type ihuanFetcher struct {
	httpUrl       string
	statisticsUrl string
	keyUrl        string
	key           string
	zone          string
	statistics    string
	eachFetchNum  int
	timeout       time.Duration
	client        *fasthttp.Client
}

func newIHuanFetcher(hu, su, ku, zone, proxy string, timeO int64, efn int) *ihuanFetcher {
	if len(hu) == 0 {
		hu = "https://ip.ihuan.me/tqdl.html"
	}
	if len(su) == 0 {
		su = "https://ip.ihuan.me/ti.html"
	}
	if len(ku) == 0 {
		ku = "https://ip.ihuan.me/mouse.do"
	}
	if timeO == 0 {
		timeO = 15
	}
	if efn == 0 {
		efn = 100
	}
	if timeO == 0 {
		timeO = 15
	}
	f := &ihuanFetcher{
		httpUrl:       hu,
		statisticsUrl: su,
		keyUrl:        ku,
		eachFetchNum:  efn,
		zone:          zone,
		timeout:       time.Duration(timeO) * time.Second,
		client:        &fasthttp.Client{TLSConfig: &tls.Config{InsecureSkipVerify: true}},
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

func (f *ihuanFetcher) Fetch() []*Proxy {
	logrus.WithField("provider", f.Type()).Info("fetch")
	if len(f.statistics) == 0 {
		f.generateStatistics()
	}
	if len(f.key) == 0 {
		f.generateKey()
	}
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	postData := fmt.Sprintf("num=%d&port=&kill_port=&address=%s&kill_address=&anonymity=&type=&post=&sort=1&key=%s", f.eachFetchNum, f.zone, f.key)
	req.SetRequestURI(f.httpUrl)
	req.SetBodyString(postData)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.Set("Accept-Encoding", "br")
	req.Header.SetContentType("application/x-www-form-urlencoded")
	req.Header.SetUserAgent("Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.85 Safari/537.36")
	req.Header.SetReferer("https://ip.ihuan.me/ti.html")
	if err := f.client.DoTimeout(req, res, f.timeout); err != nil {
		logrus.WithError(err).WithField("url", f.httpUrl).Error("failed to get proxy")
		return nil
	}

	body, err := readBody(res)
	if err != nil {
		logrus.WithError(err).WithField("raw", string(res.Body())).Error("failed to unGzip body")
		return nil
	}
	Ips := ipPattern.FindAll(body, -1)
	if Ips == nil {
		logrus.WithField("raw", string(body)).Error("empty ips")
		return nil
	}

	proxies := make([]*Proxy, 0)
	for _, ip := range Ips {
		if ip == nil {
			continue
		}
		proxies = append(proxies, &Proxy{
			Address:   string(ip),
			Provider:  ProviderTypeIHuan,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			ErrTimes:  0,
			DialType:  ProxyTypeHttp,
		})
	}
	return proxies
}

func (f *ihuanFetcher) Type() string {
	return ProviderTypeIHuan
}

func (f *ihuanFetcher) generateStatistics() {
	logrus.WithField("provider", f.Type()).Info("generate statistics")
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(f.statisticsUrl)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.SetUserAgent("Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.85 Safari/537.36")
	req.Header.Set("Accept-Encoding", "br")
	if err := f.client.DoTimeout(req, res, f.timeout); err != nil {
		logrus.WithError(err).WithField("url", f.statisticsUrl).Error("failed to get statistics")
		return
	}
	if res.Header.Peek("Set-Cookie") == nil {
		logrus.WithField("raw", res.Header.String()).Error("empty statistics")
		return
	}
	f.statistics = string(res.Header.Peek("Set-Cookie"))
}

func (f *ihuanFetcher) generateKey() {
	logrus.WithField("provider", f.Type()).Info("generate key")
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(f.keyUrl)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.SetUserAgent("Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.85 Safari/537.36")
	req.Header.Set("Accept-Encoding", "br")
	req.Header.SetReferer(f.statisticsUrl)
	req.Header.Set("Cookie", f.statistics)
	if err := f.client.DoTimeout(req, res, f.timeout); err != nil {
		logrus.WithError(err).WithField("url", f.statisticsUrl).Error("failed to get statistics")
		return
	}

	bodyBytes, err := readBody(res)
	if err != nil {
		logrus.WithError(err).Error("failed to get response body bytes")
		return
	}
	key := keyPattern.Find(bodyBytes)
	if key == nil {
		logrus.WithField("raw", string(bodyBytes)).Error("empty key")
		return
	}
	f.key = string(key)
}
