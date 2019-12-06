package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/vatriathlon/stopwatch/pkg/configuration"
	"github.com/vatriathlon/stopwatch/pkg/model"
	"github.com/vatriathlon/stopwatch/pkg/server"
	"github.com/vatriathlon/stopwatch/pkg/service"
	"github.com/vatriathlon/stopwatch/testsupport"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestServer(t *testing.T) {
	config, err := configuration.New()
	require.NoError(t, err)
	suite.Run(t, &ServerTestSuite{DBTestSuite: testsupport.NewDBTestSuite(config)})
}

type ServerTestSuite struct {
	testsupport.DBTestSuite
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
		// when
		req := httptest.NewRequest(echo.GET, "http://localhost:8080/api/status", nil)
		rec := httptest.NewRecorder()
		c := s.srv.NewContext(req, rec)
		// then
		assert.NoError(t, server.Status(c))
	})
}

func (s *ServerTestSuite) TestListRaces() {

	s.T().Run("ok", func(t *testing.T) {
		// given 3 races
		raceRepo := model.NewRaceRepository(s.DB)
		for i := 0; i < 5; i++ {
			race := model.Race{
				Name: fmt.Sprintf("race %s", uuid.NewV4()),
			}
			err := raceRepo.Create(&race)
			require.NoError(t, err)
		}
		// when
		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := s.srv.NewContext(req, rec)
		c.SetPath("/api/races")
		err := server.ListRaces(s.svc)(c)
		// then
		require.NoError(t, err)
		var races interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &races)
		require.NoError(t, err)
		assert.IsType(t, races, []interface{}{})
		assert.True(t, len(races.([]interface{})) >= 5) // in case there are already other races in the DB
	})
}
func (s *ServerTestSuite) TestListTeams() {

	s.T().Run("ok", func(t *testing.T) {
		// given 3 teams in a race
		raceRepo := model.NewRaceRepository(s.DB)
		race := model.Race{
			Name: fmt.Sprintf("race %s", uuid.NewV4()),
		}
		err := raceRepo.Create(&race)
		require.NoError(t, err)
		teamRepo := model.NewTeamRepository(s.DB)
		for i := 1; i < 6; i++ {
			team := testsupport.NewTeam(race.ID, i)
			err := teamRepo.Create(&team)
			require.NoError(t, err)
		}
		require.NoError(t, err)
		// when
		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := s.srv.NewContext(req, rec)
		c.SetPath(server.ListTeamsPathTmpl)
		c.SetParamNames("raceID")
		c.SetParamValues(strconv.Itoa(race.ID))
		err = server.ListTeams(s.svc)(c)
		// then
		require.NoError(t, err)
		var teams interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &teams)
		require.NoError(t, err)
		assert.IsType(t, teams, []interface{}{})
		teams = teams.([]interface{})
		require.Len(t, teams, 5)
	})
}

func (s *ServerTestSuite) TestAddLap() {

	s.T().Run("ok", func(t *testing.T) {
		// given
		raceRepo := model.NewRaceRepository(s.DB)
		race := model.Race{
			Name: "foo",
		}
		err := raceRepo.Create(&race)
		require.NoError(t, err)
		teamRepo := model.NewTeamRepository(s.DB)
		team := testsupport.NewTeam(race.ID, 1)
		err = teamRepo.Create(&team)
		require.NoError(t, err)
		// when
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := s.srv.NewContext(req, rec)
		c.SetPath(server.AddLapPathTmpl)
		c.SetParamNames("raceID", "bibnumber")
		c.SetParamValues(strconv.Itoa(race.ID), strconv.Itoa(team.BibNumber))
		err = server.AddLap(s.svc)(c)
		// then
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
	})

}
