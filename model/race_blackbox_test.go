package model_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/vatriathlon/stopwatch/configuration"
	"github.com/vatriathlon/stopwatch/model"
	testsuite "github.com/vatriathlon/stopwatch/test/suite"
)

func TestRaceRepository(t *testing.T) {
	config, err := configuration.New()
	require.NoError(t, err)
	suite.Run(t, &RaceRepositoryTestSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

type RaceRepositoryTestSuite struct {
	testsuite.DBTestSuite
}

func (s *RaceRepositoryTestSuite) TestCreateRace() {
	// given
	ctx := context.Background()
	s.T().Run("ok", func(t *testing.T) {
		// given
		race := model.Race{
			Name: fmt.Sprintf("race %s", uuid.NewV4()),
		}
		// when
		err := s.App.Races().Create(ctx, &race)
		// then
		require.NoError(t, err)
		require.NotEqual(t, race.ID, uuid.Nil)
	})

	s.T().Run("failure", func(t *testing.T) {

		t.Run("missing name", func(t *testing.T) {
			// given
			race := model.Race{}
			// when
			err := s.App.Races().Create(ctx, &race)
			// then
			require.NoError(t, err)
			require.NotEqual(t, race.ID, uuid.Nil)
		})
	})

}

func (s *RaceRepositoryTestSuite) TestStartRace() {
	// given
	ctx := context.Background()

	s.T().Run("ok", func(t *testing.T) {
		// given
		race := model.Race{
			Name: fmt.Sprintf("race %s", uuid.NewV4()),
		}
		err := s.App.Races().Create(ctx, &race)
		require.NoError(t, err)
		// when
		err = s.App.Races().Start(ctx, &race)
		// then
		require.NoError(t, err)
		require.True(t, race.IsStarted())
	})

	s.T().Run("failure", func(t *testing.T) {

		t.Run("already started", func(t *testing.T) {
			// given
			race := model.Race{
				Name: fmt.Sprintf("race %s", uuid.NewV4()),
			}
			err := s.App.Races().Create(ctx, &race)
			require.NoError(t, err)
			err = s.App.Races().Start(ctx, &race)
			require.NoError(t, err)
			// when
			err = s.App.Races().Start(ctx, &race)
			// then
			require.Error(t, err)
		})
	})
}

func (s *RaceRepositoryTestSuite) TestEndRace() {
	// given
	ctx := context.Background()

	s.T().Run("ok", func(t *testing.T) {
		// given
		race := model.Race{
			Name: fmt.Sprintf("race %s", uuid.NewV4()),
		}
		err := s.App.Races().Create(ctx, &race)
		require.NoError(t, err)
		err = s.App.Races().Start(ctx, &race)
		require.NoError(t, err)
		// when
		err = s.App.Races().End(ctx, &race)
		// then
		require.NoError(t, err)
		require.True(t, race.IsEnded())
	})

	s.T().Run("failure", func(t *testing.T) {

		t.Run("not started yet", func(t *testing.T) {
			// given
			race := model.Race{
				Name: fmt.Sprintf("race %s", uuid.NewV4()),
			}
			err := s.App.Races().Create(ctx, &race)
			require.NoError(t, err)
			err = s.App.Races().Start(ctx, &race)
			require.NoError(t, err)
			// when
			err = s.App.Races().Start(ctx, &race)
			// then
			require.Error(t, err)
		})

		t.Run("already ended", func(t *testing.T) {
			// given
			race := model.Race{
				Name: fmt.Sprintf("race %s", uuid.NewV4()),
			}
			err := s.App.Races().Create(ctx, &race)
			require.NoError(t, err)
			err = s.App.Races().Start(ctx, &race)
			require.NoError(t, err)
			err = s.App.Races().End(ctx, &race)
			require.NoError(t, err)
			// when
			err = s.App.Races().End(ctx, &race)
			// then
			require.Error(t, err)
		})
	})

}

func (s *RaceRepositoryTestSuite) TestListRacesNoResult() {
	// given
	ctx := context.Background()
	// when
	races, err := s.App.Races().List(ctx)
	// then
	require.NoError(s.T(), err)
	assert.Empty(s.T(), races)
}

func (s *RaceRepositoryTestSuite) TestListRacesSingleResult() {
	// given
	ctx := context.Background()
	race1 := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err := s.App.Races().Create(ctx, &race1)
	require.NoError(s.T(), err)
	// when
	races, err := s.App.Races().List(ctx)
	// then
	require.NoError(s.T(), err)
	assert.Len(s.T(), races, 1)
}

func (s *RaceRepositoryTestSuite) TestListRacesMultipleResults() {
	// given
	ctx := context.Background()
	race1 := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	race2 := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err := s.App.Races().Create(ctx, &race1)
	require.NoError(s.T(), err)
	err = s.App.Races().Create(ctx, &race2)
	require.NoError(s.T(), err)
	// when
	races, err := s.App.Races().List(ctx)
	// then
	require.NoError(s.T(), err)
	assert.Len(s.T(), races, 2)
}
