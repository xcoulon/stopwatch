package model_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/vatriathlon/stopwatch/configuration"
	"github.com/vatriathlon/stopwatch/model"
	testmodel "github.com/vatriathlon/stopwatch/test/model"
	testsuite "github.com/vatriathlon/stopwatch/test/suite"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestLapRepository(t *testing.T) {
	config, err := configuration.New()
	require.NoError(t, err)
	suite.Run(t, &LapRepositoryTestSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

type LapRepositoryTestSuite struct {
	testsuite.DBTestSuite
}

func (s *LapRepositoryTestSuite) TestCreateLap() {
	// given
	raceRepo := model.NewRaceRepository(s.DB)
	teamRepo := model.NewTeamRepository(s.DB)
	lapRepo := model.NewLapRepository(s.DB)
	now := time.Now()
	race := model.Race{
		Name: fmt.Sprintf("race-%s", uuid.NewV4()),
	}
	err := raceRepo.Create(&race)
	require.NoError(s.T(), err)
	team1 := testmodel.NewTeam(race.ID, 1)
	err = teamRepo.Create(&team1)
	require.NoError(s.T(), err)
	team2 := testmodel.NewTeam(race.ID, 2)
	err = teamRepo.Create(&team2)
	require.NoError(s.T(), err)

	s.T().Run("ok", func(t *testing.T) {
		// given
		lap1 := model.Lap{
			RaceID: race.ID,
			TeamID: team1.ID,
			Time:   now.Add(1 * time.Minute),
		}
		// when
		err := lapRepo.Create(&lap1)
		// then
		require.NoError(t, err)
		require.NotEqual(t, lap1.ID, uuid.Nil)
	})

	s.T().Run("failure", func(t *testing.T) {

		t.Run("missing RaceID", func(t *testing.T) {
			// given
			lap1 := model.Lap{
				TeamID: team1.ID,
				Time:   now.Add(1 * time.Minute),
			}
			// when
			err := lapRepo.Create(&lap1)
			// then
			require.Error(t, err)
			assert.Equal(t, err.Error(), "missing 'RaceID' field")
		})

		t.Run("missing TeamID", func(t *testing.T) {
			// given
			lap1 := model.Lap{
				RaceID: race.ID,
				Time:   now.Add(1 * time.Minute),
			}
			// when
			err := lapRepo.Create(&lap1)
			// then
			require.Error(t, err)
			assert.Equal(t, err.Error(), "missing 'TeamID' field")
		})
	})
}
