package core

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"strconv"
	"strings"
)

func logger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		logrus.WithFields(logrus.Fields{
			"path":   c.Path(),
			"client": c.Request().RemoteAddr,
		}).Info("receive request")
		return next(c)
	}
}

func checkRequireProxyParam(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !strings.Contains(c.Path(), "require") {
			return next(c)
		}
		// check param number
		numberString := c.QueryParam("num")
		if len(numberString) == 0 {
			return c.JSON(http.StatusOK, map[string]any{
				"code": httpFailed,
				"err":  "missing param num",
			})
		}
		numberInt, err := strconv.Atoi(numberString)
		if err != nil {
			logrus.WithError(err).WithField("raw", numberString).Error("failed to parse string to int")
			return c.JSON(http.StatusOK, map[string]any{
				"code": httpFailed,
				"err":  fmt.Sprintf("failed to parse %s to int", numberString),
			})
		}
		if numberInt == 0 {
			numberInt = 1
		}
		if numberInt > viper.GetInt("max_get_number") {
			numberInt = 100
		}
		c.Set("num", numberInt)

		// check param provider
		provider := c.QueryParam("provider")
		if len(provider) == 0 {
			return c.JSON(http.StatusOK, map[string]any{
				"code": httpFailed,
				"err":  "missing param provider",
			})
		}
		providerSlice := strings.Split(provider, ",")
		uniqueProvider := map[string]bool{}
		for _, pvd := range providerSlice {
			exist := false
			for _, each := range UserAvailableProviderType {
				if strings.ToUpper(pvd) == each {
					exist = true
					break
				}
			}
			if !exist {
				logrus.WithField("provider", pvd).Error("unknown provider")
				continue
			}
			if strings.ToUpper(pvd) == ProviderTypeMix {
				for _, each := range AllProviderType {
					uniqueProvider[each] = true
				}
			} else {
				uniqueProvider[pvd] = true
			}
		}
		uniqueSlice := make([]string, 0)
		for pvd := range uniqueProvider {
			uniqueSlice = append(uniqueSlice, pvd)
		}
		c.Set("providers", uniqueSlice)

		return next(c)
	}
}

func checkReportErrorParam(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !strings.Contains(c.Path(), "report") {
			return next(c)
		}
		if len(c.QueryParam("id")) == 0 {
			return c.JSON(http.StatusOK, map[string]any{
				"code": httpFailed,
				"err":  "missing param id",
			})
		}
		return next(c)
	}
}
