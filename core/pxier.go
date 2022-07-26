package core

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
	"strings"
	"sync"
	"time"
)

type Pxier struct {
	*echo.Echo
	startTime int64
	maxErr    int
	fetchers  []fetcher
	db        *gorm.DB
	stop      bool
	dbCache   map[string]map[int]*Proxy
	cacheLock *sync.RWMutex
}

// NewPixer creates a new Pxier instance and return
func NewPixer() *Pxier {
	p := &Pxier{
		cacheLock: &sync.RWMutex{},
		Echo:      echo.New(),
		startTime: time.Now().Unix(),
		maxErr:    viper.GetInt("max_error_time"),
		fetchers:  make([]fetcher, 0),
		db:        newDB(),
		stop:      false,
		dbCache:   map[string]map[int]*Proxy{},
	}
	p.initFetcher()
	p.registerMiddleware()
	p.registerRoute()
	return p
}

// Run starts the Pxier fetching loop
func (p *Pxier) Run() {
	go p.syncCache()
	go p.syncDB()
	go func() {
		if len(viper.GetString("listen")) == 0 {
			logrus.Panic("missing param listen")
		}
		logrus.Fatal(p.Start(viper.GetString("listen")))
	}()

	interval := viper.GetInt64("fetch_interval")
	if interval == 0 {
		interval = 60
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
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

// syncDB sync local cache to database
func (p *Pxier) syncDB() {
	interval := viper.GetInt64("sync_interval")
	if interval == 0 {
		interval = 15
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for {
		logrus.Info("sync cache to database")
		p.cacheLock.Lock()
		for pvd, proxyMap := range p.dbCache {
			for pid, each := range proxyMap {
				go p.db.Save(&each)
				// After sync to database, check if max err
				if each.ErrTimes > p.maxErr {
					delete(p.dbCache[pvd], pid)
				}
			}
		}
		p.cacheLock.Unlock()
		<-ticker.C
	}
}

// syncCache will load database data to local cache
func (p *Pxier) syncCache() {
	interval := viper.GetInt64("sync_interval")
	if interval == 0 {
		interval = 15
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for {
		logrus.Info("sync database to cache")
		for _, pvd := range AllProviderType {
			temp := make([]*Proxy, 0)
			if err := p.db.Where("provider = ? and err_times <= ?", pvd, p.maxErr).Find(&temp).Error; err != nil {
				if strings.Contains(err.Error(), "Too many connections") {
					continue
				}
				logrus.WithError(err).Panic("failed to sync database to cache")
			}
			p.cacheLock.Lock()
			for _, each := range temp {
				if p.dbCache[pvd] == nil {
					p.dbCache[pvd] = map[int]*Proxy{}
				}
				if p.dbCache[pvd][each.Id] != nil {
					continue
				}
				p.dbCache[pvd][each.Id] = each
			}
			p.cacheLock.Unlock()
		}
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
	if len(proxies) == 0 {
		return
	}
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
	//p.Use(logger)
	p.Use(checkRequireProxyParam)
	p.Use(checkReportErrorParam)
}

// registerRoute will register routes for *echo.Echo
func (p *Pxier) registerRoute() {
	p.GET("/status", p.apiGetStatus)
	p.GET("/require", p.apiGetProxy)
	p.GET("/report", p.apiReportError)
}
