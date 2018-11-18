package model_test

import (
	"context"
	"fmt"
	"testing"

	uuid "github.com/satori/go.uuid"
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
	ctx := context.Background()
	race := model.Race{
		Name: fmt.Sprintf("team %s", uuid.NewV4()),
	}
	// when
	err := s.App.Races().Create(ctx, &race)
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
		err := s.App.Teams().Create(ctx, &team)
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
			err := s.App.Teams().Create(ctx, &team)
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
			err := s.App.Teams().Create(ctx, &team)
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
			err := s.App.Teams().Create(ctx, &team)
			// then
			require.Error(t, err)
		})
	})

}
