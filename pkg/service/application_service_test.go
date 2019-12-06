package service_test

import (
	"fmt"
	"testing"

	"github.com/vatriathlon/stopwatch/pkg/configuration"
	"github.com/vatriathlon/stopwatch/pkg/model"
	"github.com/vatriathlon/stopwatch/pkg/service"
	"github.com/vatriathlon/stopwatch/testsupport"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestAppService(t *testing.T) {
	config, err := configuration.New()
	require.NoError(t, err)
	suite.Run(t, &AppServiceTestSuite{DBTestSuite: testsupport.NewDBTestSuite(config)})
}

type AppServiceTestSuite struct {
	testsupport.DBTestSuite
}

func (s *AppServiceTestSuite) TestListRacesNoResult() {
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

func (s *AppServiceTestSuite) TestListRacesMultipleResults() {
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
func (s *AppServiceTestSuite) TestGetRace() {

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

func (s *AppServiceTestSuite) TestListTeams() {

	s.T().Run("ok", func(t *testing.T) {
		// given
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
		svc := service.NewApplicationService(s.DB)
		// when
		teams, err := svc.ListTeams(race.ID)
		// then
		require.NoError(t, err)
		assert.Len(t, teams, 5)
	})
}

func (s *AppServiceTestSuite) TestStartRace() {
	// given
	raceRepo := model.NewRaceRepository(s.DB)
	svc := service.NewApplicationService(s.DB)

	s.T().Run("ok", func(t *testing.T) {
		// given
		race := model.Race{
			Name: fmt.Sprintf("race %s", uuid.NewV4()),
		}
		err := raceRepo.Create(&race)
		require.NoError(t, err)
		// when
		_, err = svc.StartRace(race.ID)
		// then
		require.NoError(t, err)
		// verify the start time
		result, err := raceRepo.FindByName(race.Name)
		require.NoError(s.T(), err)
		require.True(t, result.IsStarted())
		assert.False(s.T(), result.StartTime.IsZero())
		assert.True(s.T(), result.EndTime.IsZero())
	})

	s.T().Run("failure", func(t *testing.T) {

		t.Run("already started", func(t *testing.T) {
			// given
			race := model.Race{
				Name: fmt.Sprintf("race %s", uuid.NewV4()),
			}
			err := raceRepo.Create(&race)
			require.NoError(t, err)
			_, err = svc.StartRace(race.ID)
			require.NoError(t, err)
			// when
			_, err = svc.StartRace(race.ID)
			// then
			require.Error(t, err)
		})
	})
}

func (s *AppServiceTestSuite) TestAddLap() {

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
	for i := 1; i < 6; i++ {
		team := testsupport.NewTeam(race.ID, i)
		err := teamRepo.Create(&team)
		require.NoError(s.T(), err)
		teams = append(teams, team)
	}

	s.T().Run("ok", func(t *testing.T) {

		t.Run("team 1 lap 1", func(t *testing.T) {
			// when
			team, err := svc.AddLap(race.ID, 1)
			// then
			require.NoError(t, err)
			require.Equal(t, teams[0].Name, team.Name)
			assert.Len(t, team.Laps, 1)
		})

		t.Run("team 2 lap 1+2", func(t *testing.T) {
			// when
			_, err := svc.AddLap(race.ID, 2)
			require.NoError(t, err)
			team, err := svc.AddLap(race.ID, 2)
			// then
			require.NoError(t, err)
			require.Equal(t, teams[1].Name, team.Name)
			assert.Len(t, team.Laps, 2)
		})
	})
}
func (s *AppServiceTestSuite) TestFirstAddLapForAll() {

	s.T().Run("enabled", func(t *testing.T) {
		// given
		raceRepo := model.NewRaceRepository(s.DB)
		race := model.Race{
			Name:           fmt.Sprintf("race %s", uuid.NewV4()),
			AllowsFirstLap: true,
			HasFirstLap:    false,
		}
		err := raceRepo.Create(&race)
		require.NoError(t, err)
		svc := service.NewApplicationService(s.DB)
		teamRepo := model.NewTeamRepository(s.DB)
		teams := []model.Team{}
		for i := 1; i < 6; i++ {
			team := testsupport.NewTeam(race.ID, i)
			err := teamRepo.Create(&team)
			require.NoError(t, err)
			teams = append(teams, team)
		}

		t.Run("can add first lap", func(t *testing.T) {
			// when
			race, err := svc.AddFirstLapForAll(race.ID)
			// then
			require.NoError(t, err)
			assert.True(t, race.AllowsFirstLap)
			assert.True(t, race.HasFirstLap)
			// check that all teams have a lap
			teams, err := teamRepo.List(race.ID)
			require.NoError(t, err)
			for _, team := range teams {
				assert.NotEmpty(t, team.Laps)
			}
		})

		s.T().Run("cannot add first lap again", func(t *testing.T) {
			// when
			_, err := svc.AddFirstLapForAll(race.ID)
			// then
			require.Error(t, err)
		})
	})

	s.T().Run("disabled", func(t *testing.T) {
		// given
		raceRepo := model.NewRaceRepository(s.DB)
		race := model.Race{
			Name:           fmt.Sprintf("race %s", uuid.NewV4()),
			AllowsFirstLap: false,
			HasFirstLap:    false,
		}
		err := raceRepo.Create(&race)
		require.NoError(t, err)
		svc := service.NewApplicationService(s.DB)
		// when
		_, err = svc.AddFirstLapForAll(race.ID)
		// then
		require.Error(t, err)
	})

}
