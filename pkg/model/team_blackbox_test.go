package model_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/vatriathlon/stopwatch/pkg/configuration"
	"github.com/vatriathlon/stopwatch/pkg/model"
	"github.com/vatriathlon/stopwatch/testsupport" 

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestTeamRepository(t *testing.T) {
	config, err := configuration.New()
	require.NoError(t, err)
	suite.Run(t, &TeamRepositoryTestSuite{DBTestSuite: testsupport.NewDBTestSuite(config)})
}

type TeamRepositoryTestSuite struct {
	testsupport.DBTestSuite
}

func (s *TeamRepositoryTestSuite) TestCreateTeam() {
	// given
	raceRepo := model.NewRaceRepository(s.DB)
	teamRepo := model.NewTeamRepository(s.DB)
	race := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	// when
	err := raceRepo.Create(&race)
	// then
	require.NoError(s.T(), err)

	s.T().Run("ok", func(t *testing.T) {
		// given
		team := testsupport.NewTeam(race.ID, 1)
		// when
		err := teamRepo.Create(&team)
		// then
		require.NoError(t, err)
		require.NotEqual(t, team.ID, uuid.Nil)
	})

	s.T().Run("failure", func(t *testing.T) {

		t.Run("missing name", func(t *testing.T) {
			// given
			team := testsupport.NewTeam(race.ID, 1)
			team.Name = ""
			// when
			err := teamRepo.Create(&team)
			// then
			require.Error(t, err)
		})

		t.Run("missing bib number", func(t *testing.T) {
			// given
			team := testsupport.NewTeam(race.ID, 0)
			// when
			err := teamRepo.Create(&team)
			// then
			require.Error(t, err)
		})

		t.Run("duplicate bib number", func(t *testing.T) {
			// given
			team1 := testsupport.NewTeam(race.ID, 2)
			err := teamRepo.Create(&team1)
			require.NoError(t, err)
			// when
			team2 := testsupport.NewTeam(race.ID, 2)
			err = teamRepo.Create(&team2)
			// then
			require.Error(t, err)
		})

		t.Run("missing race ID", func(t *testing.T) {
			// given
			team := testsupport.NewTeam(race.ID, 1)
			team.RaceID = 0
			// when
			err := teamRepo.Create(&team)
			// then
			require.Error(t, err)
		})
	})
}

func (s *TeamRepositoryTestSuite) TestListTeamsNoResult() {
	// given
	raceRepo := model.NewRaceRepository(s.DB)
	teamRepo := model.NewTeamRepository(s.DB)
	race := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err := raceRepo.Create(&race)
	require.NoError(s.T(), err)
	// when
	teams, err := teamRepo.List(race.ID)
	// then
	require.NoError(s.T(), err)
	assert.Empty(s.T(), teams)
}

func (s *TeamRepositoryTestSuite) TestListTeamsSingleResult() {
	// given
	raceRepo := model.NewRaceRepository(s.DB)
	teamRepo := model.NewTeamRepository(s.DB)
	race := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err := raceRepo.Create(&race)
	require.NoError(s.T(), err)
	team := testsupport.NewTeam(race.ID, 1)
	err = teamRepo.Create(&team)
	require.NoError(s.T(), err)
	// when
	teams, err := teamRepo.List(race.ID)
	// then
	require.NoError(s.T(), err)
	require.Len(s.T(), teams, 1)
	assert.Equal(s.T(), team.ID, teams[0].ID)
}

func (s *TeamRepositoryTestSuite) TestListTeamsMultipleResults() {
	// given
	raceRepo := model.NewRaceRepository(s.DB)
	teamRepo := model.NewTeamRepository(s.DB)
	lapRepo := model.NewLapRepository(s.DB)
	race := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err := raceRepo.Create(&race)
	require.NoError(s.T(), err)
	team1 := testsupport.NewTeam(race.ID, 2)
	err = teamRepo.Create(&team1)
	require.NoError(s.T(), err)
	lap1 := model.Lap{
		RaceID: race.ID,
		TeamID: team1.ID,
		Time:   time.Now(),
	}
	err = lapRepo.Create(&lap1)
	require.NoError(s.T(), err)
	team2 := testsupport.NewTeam(race.ID, 1)
	err = teamRepo.Create(&team2)
	require.NoError(s.T(), err)
	// when
	teams, err := teamRepo.List(race.ID)
	// then
	require.NoError(s.T(), err)
	require.Len(s.T(), teams, 2)
	assert.Equal(s.T(), team2.ID, teams[0].ID)
	assert.Equal(s.T(), team1.ID, teams[1].ID)
	assert.Len(s.T(), teams[1].Laps, 1)
}
