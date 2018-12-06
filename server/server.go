package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/vatriathlon/stopwatch/configuration"
	"github.com/vatriathlon/stopwatch/service"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
)

// New instanciates a new Echo server
func New(svc service.ApplicationService) *echo.Echo {
	// starts the HTTP engine to handle requests
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// graceful handle of errors, i.e., just logging with the same logger as everywhere else in the app.
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if he, ok := err.(*echo.HTTPError); ok {
			logrus.WithField("code", he.Code).WithField("request_url", c.Request().RequestURI).Error(he.Message)
			if msg, ok := he.Message.(string); ok {
				c.String(he.Code, msg)
			} else {
				c.NoContent(he.Code)
			}
		}
	}
	e.POST("/api/races/:race/teams/:team", AddLap(svc))
	e.GET("/api/status", Status)
	return e
}

// Status returns a basic `ping/pong` handler
func Status(c echo.Context) error {
	return c.String(http.StatusOK, fmt.Sprintf("build.time: %s - build.commit: %s 👷‍♂️", configuration.BuildTime, configuration.BuildCommit))
}

// AddLap returns a handler to create an db record for the team in the race
func AddLap(svc service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		scheme := c.Scheme()
		host := c.Request().Host
		logrus.Debugf("Processing incoming request on %s://%s%s", scheme, host, c.Request().URL)
		raceID, err := strconv.Atoi(c.Param("raceID"))
		if err != nil {
			return err
		}
		bibnumber := c.Param("bibNumber")
		team, err := svc.AddLap(raceID, bibnumber)
		if err != nil {
			return err
		}
		c.JSON(http.StatusCreated, team)
		return nil
	}
}
