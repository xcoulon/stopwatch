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
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
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
	e.GET("/api/status", Status)
	e.GET("/api/races", ListRaces(svc))
	e.GET("/api/races/:raceID", ShowRace(svc))
	e.GET("/api/races/:raceID/teams", ListTeams(svc))
	e.POST("/api/races/:raceID/bibnumber/:bibnumber/laps", AddLap(svc))
	return e
}

// Status returns a basic `ping/pong` handler
func Status(c echo.Context) error {
	return c.String(http.StatusOK, fmt.Sprintf("build.time: %s - build.commit: %s 👷‍♂️", configuration.BuildTime, configuration.BuildCommit))
}

// ShowRace returns a handler to list races
func ShowRace(svc service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		scheme := c.Scheme()
		host := c.Request().Host
		logrus.Debugf("Processing incoming request on %s://%s%s", scheme, host, c.Request().URL)
		raceID, err := strconv.Atoi(c.Param("raceID"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("unable to convert race id '%s' to integer", c.Param("raceID")))
		}
		race, err := svc.GetRace(raceID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, race)
	}
}

// ListRaces returns a handler to list races
func ListRaces(svc service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		scheme := c.Scheme()
		host := c.Request().Host
		logrus.Debugf("Processing incoming request on %s://%s%s", scheme, host, c.Request().URL)
		races, err := svc.ListRaces()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, races)
	}
}

// ListTeams returns a handler to list teams in a given race
func ListTeams(svc service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		scheme := c.Scheme()
		host := c.Request().Host
		logrus.Debugf("Processing incoming request on %s://%s%s", scheme, host, c.Request().URL)
		raceID, err := strconv.Atoi(c.Param("raceID"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("unable to convert race id '%s' to integer", c.Param("raceID")))
		}
		teams, err := svc.ListTeams(raceID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, teams)
	}
}

// AddLap returns a handler to create an db record for the team in the race
func AddLap(svc service.ApplicationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		scheme := c.Scheme()
		host := c.Request().Host
		logrus.Debugf("Processing incoming request on %s://%s%s", scheme, host, c.Request().URL)
		raceID, err := strconv.Atoi(c.Param("raceID"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("unable to convert race id '%s' to integer", c.Param("raceID")))
		}
		bibnumber := c.Param("bibnumber")
		team, err := svc.AddLap(raceID, bibnumber)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusCreated, team)
	}
}
