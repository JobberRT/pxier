package core

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"time"
)

func (p *Pxier) apiGetStatus(c echo.Context) error {
	result := map[string]any{}
	runningDuration := time.Now().Sub(time.Unix(p.startTime, 0)).String()
	result["running"] = runningDuration
	for _, pvd := range AllProviderType {
		var count int64
		if err := p.db.Model(&Proxy{}).Where(&Proxy{Provider: pvd}).Count(&count); err.Error != nil {
			logrus.WithError(err.Error).WithField("provider", pvd).Error("failed to query db")
			return c.JSON(http.StatusOK, map[string]any{
				"code": httpFailed,
				"err":  err.Error,
			})
		}
		result[pvd] = count
	}
	return c.JSON(http.StatusOK, map[string]any{
		"code": httpSuccess,
		"data": result,
	})
}

func (p *Pxier) apiGetProxy(c echo.Context) error {
	num := c.Get("num").(int)
	providers := c.Get("providers").([]string)
	if num > len(providers) {
		providers = make([]string, 0)
		for len(providers) < num {
			providers = append(providers, AllProviderType[rand.Intn(len(AllProviderType))])
		}
	}
	eachProviderNum := num / len(providers)
	res := make([]*Proxy, 0)
	for _, pvd := range providers {
		if len(res) >= num {
			break
		}
		temp := make([]*Proxy, 0)
		p.db.Raw("select * from proxy where provider = UPPER(?) order by RAND() limit ?", pvd, eachProviderNum).Scan(&temp)
		res = append(res, temp...)
	}
	return c.JSON(http.StatusOK, map[string]any{
		"code": httpSuccess,
		"data": res,
	})
}

func (p *Pxier) apiReportError(c echo.Context) error {
	address := c.QueryParam("address")
	temp := &Proxy{}
	if err := p.db.Where(&Proxy{Address: address}).First(&temp); err.Error != nil {
		logrus.WithError(err.Error).WithField("address", address).Error("unknown address")
		return c.JSON(http.StatusOK, map[string]any{
			"code": httpFailed,
			"err":  fmt.Sprintf("unknown address: %s", address),
		})
	}

	p.db.Model(&temp).Update("err_times", temp.ErrTimes+1)
	return c.JSON(http.StatusOK, map[string]any{
		"code": httpSuccess,
		"data": "success",
	})
}
