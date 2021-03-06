package core

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

func (p *Pxier) apiGetStatus(c echo.Context) error {
	result := map[string]any{}
	runningDuration := time.Now().Sub(time.Unix(p.startTime, 0)).String()
	result["running"] = runningDuration
	result["data"] = p.dbCache.Items()
	result["total"] = p.dbCache.Count()
	return c.JSON(http.StatusOK, map[string]any{
		"code": httpSuccess,
		"data": result,
	})
}

func (p *Pxier) apiGetProxy(c echo.Context) error {
	num := c.Get("num").(int)
	providers := c.Get("providers").([]string)

	res := make([]*Proxy, 0)
	for _, pxy := range p.dbCache.Items() {
		if pxy.ErrTimes > p.maxErr {
			continue
		}
		for _, pvd := range providers {
			if pxy.Provider == pvd {
				res = append(res, pxy)
			}
		}
		if len(res) >= num {
			break
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"code": httpSuccess,
		"data": res,
	})
}

func (p *Pxier) apiReportError(c echo.Context) error {
	id := c.QueryParam("id")
	pxy, ok := p.dbCache.Get(id)
	if !ok {
		return c.JSON(http.StatusOK, map[string]any{
			"code": httpFailed,
			"data": fmt.Sprintf("no such proxy id: %s", id),
		})
	}
	pxy.ErrTimes++
	p.dbCache.Set(id, pxy)
	return c.JSON(http.StatusOK, map[string]any{
		"code": httpSuccess,
		"data": "success",
	})
}
