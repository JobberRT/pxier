package core

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
	"strings"
	"time"
)

type Pxier struct {
	startTime int64
	maxErr    int
	*echo.Echo
	fetchers []fetcher
	db       *gorm.DB
	stop     bool
}

// NewPixer creates a new Pxier instance and return
func NewPixer() *Pxier {
	p := &Pxier{
		Echo:      echo.New(),
		startTime: time.Now().Unix(),
		maxErr:    viper.GetInt("max_error_time"),
		fetchers:  make([]fetcher, 0),
		db:        newDB(),
		stop:      false,
	}
	p.initFetcher()
	p.registerMiddleware()
	p.registerRoute()
	return p
}

// Run starts the Pxier fetching loop
func (p *Pxier) Run() {
	go func() {
		go p.removeDeadProxy()
		if len(viper.GetString("listen")) == 0 {
			logrus.Panic("missing param listen")
		}
		logrus.Fatal(p.Start(viper.GetString("listen")))
	}()
	ticker := time.NewTicker(time.Duration(viper.GetInt64("fetch_interval")) * time.Second)
	defer ticker.Stop()
	for {
		if p.stop {
			return
		}
		for _, f := range p.fetchers {
			ft := f
			go func() {
				proxies := ft.Fetch()
				p.insertProxy(proxies)
			}()
		}
		<-ticker.C
	}
}

// Stop stops the Pxier fetching loop
func (p *Pxier) Stop() {
	p.stop = true
	_ = p.Close()
}

// removeDeadProxy removes all the proxies that exceed the max error time
func (p *Pxier) removeDeadProxy() {
	logrus.Info("remove dead proxy")
	interval := viper.GetInt64("remove_interval")
	if interval == 0 {
		logrus.Warn("remove_interval is 0, set to 60")
		interval = 60
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for {
		if p.stop {
			return
		}
		p.db.Delete(&Proxy{}, "err_times > ?", fmt.Sprintf("%%%d%%", p.maxErr))
		<-ticker.C
	}
}

// initFetcher creates all the fetcher
func (p *Pxier) initFetcher() {
	logrus.Info("init fetcher")
	selectedProviders := viper.GetStringSlice("providers")
	if len(selectedProviders) == 0 {
		logrus.Warn("providers is empty, set to all")
		selectedProviders = AllProviderType
	}
	for _, pvd := range selectedProviders {
		p.fetchers = append(p.fetchers, newFetcher(strings.ToUpper(pvd)))
	}
}

// insertProxy insert all proxy to the database
func (p *Pxier) insertProxy(proxies []*Proxy) {
	logrus.WithFields(logrus.Fields{
		"number":   len(proxies),
		"provider": proxies[0].Provider,
	}).Info("insert proxy")
	for _, each := range proxies {
		// Update or Create
		if p.db.Model(&Proxy{}).Where("address = ? and dial_type = ?", each.Address, each.DialType).Update("updated_at", time.Now().Unix()).RowsAffected == 0 {
			each.ErrTimes = 0
			each.CreatedAt = time.Now().Unix()
			each.UpdatedAt = time.Now().Unix()
			p.db.Create(&each)
		}
	}
}

// registerMiddleware will register needed middlewares for *echo.Echo
func (p *Pxier) registerMiddleware() {
	rateLimit := viper.GetInt("rate_limit")
	if rateLimit == 0 {
		logrus.Warn("rate_limit is 0, set to 3")
		rateLimit = 3
	}
	p.Use(middleware.Recover())
	p.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 9}))
	p.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(rateLimit))))
	p.Use(logger)
	p.Use(checkRequireProxyParam)
	p.Use(checkReportErrorParam)
}

// registerRoute will register routes for *echo.Echo
func (p *Pxier) registerRoute() {
	p.GET("/status", p.apiGetStatus)
	p.GET("/require", p.apiGetProxy)
	p.GET("/report", p.apiReportError)
}
