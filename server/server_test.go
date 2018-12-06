package server_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/vatriathlon/stopwatch/model"

	"github.com/vatriathlon/stopwatch/configuration"
	"github.com/vatriathlon/stopwatch/server"
	"github.com/vatriathlon/stopwatch/service"
	testsuite "github.com/vatriathlon/stopwatch/test/suite"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestServer(t *testing.T) {
	config, err := configuration.New()
	require.NoError(t, err)
	suite.Run(t, &ServerTestSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

type ServerTestSuite struct {
	testsuite.DBTestSuite
	config configuration.Configuration
	db     *gorm.DB
	svc    service.ApplicationService
	srv    *echo.Echo
}

func (s *ServerTestSuite) SetupTest() {
	s.DBTestSuite.SetupTest()
	s.svc = service.NewApplicationService(s.DB)
	s.srv = server.New(s.svc)
}

func (s *ServerTestSuite) TestStatusEndpoint() {

	s.T().Run("ok", func(t *testing.T) {
		// given
		req := httptest.NewRequest(echo.GET, "http://localhost:8080/api/status", nil)
		rec := httptest.NewRecorder()
		c := s.srv.NewContext(req, rec)
		// Assertions
		assert.NoError(t, server.Status(c))
	})
}

func (s *ServerTestSuite) TestAddLap() {

	s.T().Run("ok", func(t *testing.T) {
		// given
		raceRepo := model.NewRaceRepository(s.DB)
		race := model.Race{
			Name: "foo",
		}
		raceRepo.Create(&race)
		teamRepo := model.NewTeamRepository(s.DB)
		team := model.Team{
			Name:      "foo",
			BibNumber: "1",
			RaceID:    race.ID,
		}
		teamRepo.Create(&team)

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := s.srv.NewContext(req, rec)
		c.SetPath("/api/races/:raceID/teams/:bibNumber")
		c.SetParamNames("raceID", "bibNumber")
		c.SetParamValues(strconv.Itoa(race.ID), team.BibNumber)
		h := server.AddLap(s.svc)

		// Assertions
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
		}

	})

}
