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

func TestRaceService(t *testing.T) {
	config, err := configuration.New()
	require.NoError(t, err)
	suite.Run(t, &RaceServiceTestSuite{DBTestSuite: testsupport.NewDBTestSuite(config)})
}

type RaceServiceTestSuite struct {
	testsupport.DBTestSuite
}

func (s *RaceServiceTestSuite) TestListRacesNoResult() {
	// given
	raceRepo := model.NewRaceRepository(s.DB)
	race1 := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err := raceRepo.Create(&race1)
	require.NoError(s.T(), err)
	svc := service.NewRaceService(s.DB)
	// when
	races, err := svc.ListRaces()
	// then
	require.NoError(s.T(), err)
	assert.Len(s.T(), races, 1)
}

func (s *RaceServiceTestSuite) TestListRacesMultipleResults() {
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
	svc := service.NewRaceService(s.DB)
	// when
	races, err := svc.ListRaces()
	// then
	require.NoError(s.T(), err)
	assert.Len(s.T(), races, 2)
}
func (s *RaceServiceTestSuite) TestGetRace() {

	s.T().Run("ok", func(t *testing.T) {
		// given
		raceRepo := model.NewRaceRepository(s.DB)
		race := model.Race{
			Name: fmt.Sprintf("race %s", uuid.NewV4()),
		}
		err := raceRepo.Create(&race)
		require.NoError(t, err)
		svc := service.NewRaceService(s.DB)
		// when
		result, err := svc.GetRace(race.ID)
		// then
		require.NoError(t, err)
		assert.Equal(t, race.Name, result.Name)
	})

	s.T().Run("not found", func(t *testing.T) {
		// given
		svc := service.NewRaceService(s.DB)
		// when
		_, err := svc.GetRace(-1)
		// then
		require.Error(t, err)
	})

}

func (s *RaceServiceTestSuite) TestListTeams() {

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
		svc := service.NewRaceService(s.DB)
		// when
		teams, err := svc.ListTeams(race.ID)
		// then
		require.NoError(t, err)
		assert.Len(t, teams, 5)
	})
}

func (s *RaceServiceTestSuite) TestStartRace() {
	// given
	raceRepo := model.NewRaceRepository(s.DB)
	svc := service.NewRaceService(s.DB)

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

func (s *RaceServiceTestSuite) TestAddLap() {

	// given
	raceRepo := model.NewRaceRepository(s.DB)
	race := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err := raceRepo.Create(&race)
	require.NoError(s.T(), err)
	svc := service.NewRaceService(s.DB)
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
