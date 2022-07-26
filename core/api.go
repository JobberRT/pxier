package core

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func (p *Pxier) apiGetStatus(c echo.Context) error {
	result := map[string]any{}
	runningDuration := time.Now().Sub(time.Unix(p.startTime, 0)).String()
	result["running"] = runningDuration
	p.cacheLock.RLock()
	result["data"] = p.dbCache
	p.cacheLock.RUnlock()
	return c.JSON(http.StatusOK, map[string]any{
		"code": httpSuccess,
		"data": result,
	})
}

func (p *Pxier) apiGetProxy(c echo.Context) error {
	start := time.Now()
	defer func() {
		if time.Now().Sub(start) > 100*time.Millisecond {
			logrus.WithField("cost", time.Now().Sub(start).String()).Warn("getProxy time cost")
		}
	}()
	num := c.Get("num").(int)
	providers := c.Get("providers").([]string)
	eachProviderNum := num / len(providers)
	if eachProviderNum == 0 {
		eachProviderNum = 1
	}

	res := make([]*Proxy, 0)
	pidMap := map[string][]int{}
	p.cacheLock.RLock()
	for pvd, proxyMap := range p.dbCache {
		for pid, _ := range proxyMap {
			if pidMap[pvd] == nil {
				pidMap[pvd] = make([]int, 0)
			}
			pidMap[pvd] = append(pidMap[pvd], pid)
		}
	}
	for len(res) < num {
		randomProvider := providers[rand.Intn(len(providers))]
		pidSlice := pidMap[randomProvider]
		if len(pidSlice) <= 0 {
			continue
		}
		temp := p.dbCache[randomProvider][pidSlice[rand.Intn(len(pidSlice))]]
		if temp.ErrTimes > p.maxErr {
			continue
		}
		res = append(res, temp)
	}
	p.cacheLock.RUnlock()
	return c.JSON(http.StatusOK, map[string]any{
		"code": httpSuccess,
		"data": res,
	})
}

func (p *Pxier) apiReportError(c echo.Context) error {
	start := time.Now()
	defer func() {
		if time.Now().Sub(start) > 100*time.Millisecond {
			logrus.WithField("cost", time.Now().Sub(start).String()).Warn("report time cost")
		}
	}()
	id, err := strconv.Atoi(c.QueryParam("id"))
	if err != nil {
		return c.JSON(http.StatusOK, map[string]any{
			"code": httpFailed,
			"err":  fmt.Sprintf("unknown id: %s", c.QueryParam("id")),
		})
	}
	pvd := c.QueryParam("provider")
	p.cacheLock.Lock()
	if p.dbCache[pvd][id] != nil {
		p.dbCache[pvd][id].ErrTimes++
	}
	p.cacheLock.Unlock()
	return c.JSON(http.StatusOK, map[string]any{
		"code": httpSuccess,
		"data": "success",
	})
}
