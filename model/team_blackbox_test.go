package model_test

import (
	"fmt"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/vatriathlon/stopwatch/configuration"
	"github.com/vatriathlon/stopwatch/model"
	testsuite "github.com/vatriathlon/stopwatch/test/suite"
)

func TestTeamRepository(t *testing.T) {
	config, err := configuration.New()
	require.NoError(t, err)
	suite.Run(t, &TeamRepositoryTestSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

type TeamRepositoryTestSuite struct {
	testsuite.DBTestSuite
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
		number := uuid.NewV4().String()
		team := model.Team{
			Name:      fmt.Sprintf("team %s", number),
			BibNumber: number,
			RaceID:    race.ID,
		}
		// when
		err := teamRepo.Create(&team)
		// then
		require.NoError(t, err)
		require.NotEqual(t, team.ID, uuid.Nil)
	})

	s.T().Run("failure", func(t *testing.T) {

		t.Run("missing name", func(t *testing.T) {
			// given
			number := uuid.NewV4().String()
			team := model.Team{
				Name:      "",
				BibNumber: number,
				RaceID:    race.ID,
			}
			// when
			err := teamRepo.Create(&team)
			// then
			require.Error(t, err)
		})

		t.Run("missing bib number", func(t *testing.T) {
			// given
			number := uuid.NewV4().String()
			team := model.Team{
				Name:      fmt.Sprintf("team %s", number),
				BibNumber: "",
				RaceID:    race.ID,
			}
			// when
			err := teamRepo.Create(&team)
			// then
			require.Error(t, err)
		})

		t.Run("duplicate bib number", func(t *testing.T) {
			// given
			number := uuid.NewV4().String()
			team1 := model.Team{
				Name:      fmt.Sprintf("team foo %s", number),
				BibNumber: number,
				RaceID:    race.ID,
			}
			err := teamRepo.Create(&team1)
			require.NoError(t, err)
			// when
			team2 := model.Team{
				Name:      fmt.Sprintf("team bar %s", number),
				BibNumber: number,
				RaceID:    race.ID,
			}
			err = teamRepo.Create(&team2)
			// then
			require.Error(t, err)
		})

		t.Run("missing race ID", func(t *testing.T) {
			// given
			number := uuid.NewV4().String()
			team := model.Team{
				Name:      fmt.Sprintf("team %s", number),
				BibNumber: number,
			}
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
	bibNumber := uuid.NewV4().String()
	team := model.Team{
		Name:      fmt.Sprintf("team foo %s", bibNumber),
		BibNumber: bibNumber,
		RaceID:    race.ID,
	}
	err = teamRepo.Create(&team)
	require.NoError(s.T(), err)
	// when
	teams, err := teamRepo.List(race.ID)
	// then
	require.NoError(s.T(), err)
	require.Len(s.T(), teams, 1)
	assert.Equal(s.T(), team, teams[0])
}

func (s *TeamRepositoryTestSuite) TestListTeamsMultipleResults() {
	// given
	raceRepo := model.NewRaceRepository(s.DB)
	teamRepo := model.NewTeamRepository(s.DB)
	race := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err := raceRepo.Create(&race)
	require.NoError(s.T(), err)
	bibNumber1 := fmt.Sprintf("2 %s", uuid.NewV4().String())
	team1 := model.Team{
		Name:      fmt.Sprintf("team foo %s", bibNumber1),
		BibNumber: bibNumber1,
		RaceID:    race.ID,
	}
	err = teamRepo.Create(&team1)
	require.NoError(s.T(), err)
	bibNumber2 := fmt.Sprintf("1 %s", uuid.NewV4().String())
	team2 := model.Team{
		Name:      fmt.Sprintf("team bar %s", bibNumber2),
		BibNumber: bibNumber2,
		RaceID:    race.ID,
	}
	err = teamRepo.Create(&team2)
	require.NoError(s.T(), err)
	// when
	teams, err := teamRepo.List(race.ID)
	// then
	require.NoError(s.T(), err)
	require.Len(s.T(), teams, 2)
	assert.Equal(s.T(), team2, teams[0])
	assert.Equal(s.T(), team1, teams[1])
}
