package core

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
)

// fetcher is the proxy-fetching instance
type fetcher interface {
	Fetch() []*Proxy
	Type() string
}

func newFetcher(pvd string) fetcher {
	switch pvd {
	case ProviderTypeCPL:
		return newCPLFetcher(viper.GetString("provider_apis.cpl.url"), viper.GetString("fetch_proxy"), viper.GetInt64("provider_apis.cpl.timeout"))
	case ProviderTypeSTRPL:
		return newSTRFetcher(viper.GetString("provider_apis.str.http_url"), viper.GetString("provider_apis.str.socks5_url"), viper.GetString("fetch_proxy"), viper.GetInt64("provider_apis.str.timeout"))
	case ProviderTypeTSXPL:
		return newTSXFetcher(viper.GetString("provider_apis.tsx.http_url"), viper.GetString("provider_apis.tsx.socks5_url"), viper.GetString("fetch_proxy"), viper.GetInt64("provider_apis.tsx.timeout"))
	default:
		logrus.WithField("provider", pvd).Panic("unknown provider type")
		return nil
	}
}

func readBody(res *fasthttp.Response) ([]byte, error) {
	switch string(res.Header.ContentEncoding()) {
	case "gzip":
		return res.BodyGunzip()
	case "deflate":
		return res.BodyInflate()
	case "br":
		return res.BodyUnbrotli()
	default:
		return res.Body(), nil
	}
}
