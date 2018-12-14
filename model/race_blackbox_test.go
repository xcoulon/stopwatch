package model_test

import (
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
	raceRepo := model.NewRaceRepository(s.DB)

	s.T().Run("ok", func(t *testing.T) {
		// given
		race := model.Race{
			Name: fmt.Sprintf("race %s", uuid.NewV4()),
		}
		// when
		err := raceRepo.Create(&race)
		// then
		require.NoError(t, err)
		require.NotEqual(t, race.ID, 0)
	})

	s.T().Run("failure", func(t *testing.T) {

		t.Run("missing name", func(t *testing.T) {
			// given
			race := model.Race{}
			// when
			err := raceRepo.Create(&race)
			// then
			require.Error(t, err)
			require.Equal(t, race.ID, 0)
		})
	})

}

// func (s *RaceRepositoryTestSuite) TestEndRace() {
// 	// given
// 	raceRepo := model.NewRaceRepository(s.DB)

// 	s.T().Run("ok", func(t *testing.T) {
// 		// given
// 		race := model.Race{
// 			Name: fmt.Sprintf("race %s", uuid.NewV4()),
// 		}
// 		err := raceRepo.Create(&race)
// 		require.NoError(t, err)
// 		err = raceRepo.Start(&race)
// 		require.NoError(t, err)
// 		// when
// 		err = raceRepo.End(&race)
// 		// then
// 		require.NoError(t, err)
// 		require.True(t, race.IsEnded())
// 		// verify the end time
// 		result, err := raceRepo.FindByName(race.Name)
// 		require.NoError(s.T(), err)
// 		assert.False(s.T(), result.StartTime.IsZero())
// 		assert.False(s.T(), result.EndTime.IsZero())
// 	})

// 	s.T().Run("failure", func(t *testing.T) {

// 		t.Run("not started yet", func(t *testing.T) {
// 			// given
// 			race := model.Race{
// 				Name: fmt.Sprintf("race %s", uuid.NewV4()),
// 			}
// 			err := raceRepo.Create(&race)
// 			require.NoError(t, err)
// 			err = raceRepo.Start(&race)
// 			require.NoError(t, err)
// 			// when
// 			err = raceRepo.Start(&race)
// 			// then
// 			require.Error(t, err)
// 		})

// 		t.Run("already ended", func(t *testing.T) {
// 			// given
// 			race := model.Race{
// 				Name: fmt.Sprintf("race %s", uuid.NewV4()),
// 			}
// 			err := raceRepo.Create(&race)
// 			require.NoError(t, err)
// 			err = raceRepo.Start(&race)
// 			require.NoError(t, err)
// 			err = raceRepo.End(&race)
// 			require.NoError(t, err)
// 			// when
// 			err = raceRepo.End(&race)
// 			// then
// 			require.Error(t, err)
// 		})
// 	})
// }

func (s *RaceRepositoryTestSuite) TestFindByName() {
	// given
	raceRepo := model.NewRaceRepository(s.DB)

	s.T().Run("ok", func(t *testing.T) {
		// given
		race := model.Race{
			Name: fmt.Sprintf("race %s", uuid.NewV4()),
		}
		err := raceRepo.Create(&race)
		require.NoError(t, err)
		// when
		_, err = raceRepo.FindByName(race.Name)
		// then
		require.NoError(t, err)
	})

	s.T().Run("no match", func(t *testing.T) {
		// when
		_, err := raceRepo.FindByName("foo")
		// then
		require.Error(t, err)
	})
}

func (s *RaceRepositoryTestSuite) TestLookup() {
	// given
	raceRepo := model.NewRaceRepository(s.DB)

	s.T().Run("ok", func(t *testing.T) {
		// given
		race := model.Race{
			Name: fmt.Sprintf("race %s", uuid.NewV4()),
		}
		err := raceRepo.Create(&race)
		require.NoError(t, err)
		// when
		_, err = raceRepo.Lookup(race.ID)
		// then
		require.NoError(t, err)
	})

	s.T().Run("no match", func(t *testing.T) {
		// when
		_, err := raceRepo.Lookup(0)
		// then
		require.Error(t, err)
	})
}

func (s *RaceRepositoryTestSuite) TestListRacesNoResult() {
	// given
	raceRepo := model.NewRaceRepository(s.DB)
	// when
	races, err := raceRepo.List()
	// then
	require.NoError(s.T(), err)
	assert.Empty(s.T(), races)
}

func (s *RaceRepositoryTestSuite) TestListRacesSingleResult() {
	// given
	raceRepo := model.NewRaceRepository(s.DB)
	race1 := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err := raceRepo.Create(&race1)
	require.NoError(s.T(), err)
	// when
	races, err := raceRepo.List()
	// then
	require.NoError(s.T(), err)
	assert.Len(s.T(), races, 1)
}

func (s *RaceRepositoryTestSuite) TestListRacesMultipleResults() {
	// given
	raceRepo := model.NewRaceRepository(s.DB)
	race1 := model.Race{
		Name: fmt.Sprintf("race foo %s", uuid.NewV4()),
	}
	race2 := model.Race{
		Name: fmt.Sprintf("race bar %s", uuid.NewV4()),
	}
	err := raceRepo.Create(&race1)
	require.NoError(s.T(), err)
	err = raceRepo.Create(&race2)
	require.NoError(s.T(), err)
	// when
	races, err := raceRepo.List()
	// then
	require.NoError(s.T(), err)
	require.Len(s.T(), races, 2)
	// verify result ordering
	assert.Equal(s.T(), race2.Name, races[0].Name)
	assert.Equal(s.T(), race1.Name, races[1].Name)
}
