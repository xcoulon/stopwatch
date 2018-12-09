package service_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/vatriathlon/stopwatch/configuration"
	"github.com/vatriathlon/stopwatch/model"
	"github.com/vatriathlon/stopwatch/service"
	testmodel "github.com/vatriathlon/stopwatch/test/model"
	testsuite "github.com/vatriathlon/stopwatch/test/suite"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestService(t *testing.T) {
	config, err := configuration.New()
	require.NoError(t, err)
	suite.Run(t, &ServiceTestSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

type ServiceTestSuite struct {
	testsuite.DBTestSuite
}

func (s *ServiceTestSuite) TestListRacesNoResult() {
	// given
	raceRepo := model.NewRaceRepository(s.DB)
	race1 := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err := raceRepo.Create(&race1)
	require.NoError(s.T(), err)
	svc := service.NewApplicationService(s.DB)
	// when
	races, err := svc.ListRaces()
	// then
	require.NoError(s.T(), err)
	assert.Len(s.T(), races, 1)
}

func (s *ServiceTestSuite) TestListRacesMultipleResults() {
	// given
	raceRepo := model.NewRaceRepository(s.DB)
	race1 := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err := raceRepo.Create(&race1)
	require.NoError(s.T(), err)
	race2 := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err = raceRepo.Create(&race2)
	require.NoError(s.T(), err)
	svc := service.NewApplicationService(s.DB)
	// when
	races, err := svc.ListRaces()
	// then
	require.NoError(s.T(), err)
	assert.Len(s.T(), races, 2)
}
func (s *ServiceTestSuite) TestGetRace() {

	s.T().Run("ok", func(t *testing.T) {
		// given
		raceRepo := model.NewRaceRepository(s.DB)
		race := model.Race{
			Name: fmt.Sprintf("race %s", uuid.NewV4()),
		}
		err := raceRepo.Create(&race)
		require.NoError(t, err)
		svc := service.NewApplicationService(s.DB)
		// when
		result, err := svc.GetRace(race.ID)
		// then
		require.NoError(t, err)
		assert.Equal(t, race.Name, result.Name)
	})

	s.T().Run("not found", func(t *testing.T) {
		// given
		svc := service.NewApplicationService(s.DB)
		// when
		_, err := svc.GetRace(-1)
		// then
		require.Error(t, err)
	})

}

func (s *ServiceTestSuite) TestListTeams() {

	s.T().Run("ok", func(t *testing.T) {
		// given
		raceRepo := model.NewRaceRepository(s.DB)
		race := model.Race{
			Name: fmt.Sprintf("race %s", uuid.NewV4()),
		}
		err := raceRepo.Create(&race)
		require.NoError(t, err)
		teamRepo := model.NewTeamRepository(s.DB)
		for i := 0; i < 5; i++ {
			team := testmodel.NewTeam(race.ID, strconv.Itoa(i))
			err := teamRepo.Create(&team)
			require.NoError(t, err)
		}
		require.NoError(t, err)
		svc := service.NewApplicationService(s.DB)
		// when
		teams, err := svc.ListTeams(race.ID)
		// then
		require.NoError(t, err)
		assert.Len(t, teams, 5)
	})
}

func (s *ServiceTestSuite) TestAddLap() {

	// given
	raceRepo := model.NewRaceRepository(s.DB)
	race := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err := raceRepo.Create(&race)
	require.NoError(s.T(), err)
	svc := service.NewApplicationService(s.DB)
	teamRepo := model.NewTeamRepository(s.DB)
	teams := []model.Team{}
	for i := 0; i < 5; i++ {
		team := testmodel.NewTeam(race.ID, strconv.Itoa(i))
		err := teamRepo.Create(&team)
		require.NoError(s.T(), err)
		teams = append(teams, team)
	}

	s.T().Run("ok", func(t *testing.T) {

		t.Run("team 0 lap 1", func(t *testing.T) {
			// when
			team, err := svc.AddLap(race.ID, "0")
			// then
			require.NoError(t, err)
			require.Equal(t, teams[0].Name, team.Name)
			assert.Len(t, team.Laps, 1)
		})

		t.Run("team 1 lap 2", func(t *testing.T) {
			// when
			_, err := svc.AddLap(race.ID, "1")
			require.NoError(t, err)
			team, err := svc.AddLap(race.ID, "1")
			// then
			require.NoError(t, err)
			require.Equal(t, teams[1].Name, team.Name)
			assert.Len(t, team.Laps, 2)
		})
	})
}
