package model_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/vatriathlon/stopwatch/model"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/vatriathlon/stopwatch/application"
	"github.com/vatriathlon/stopwatch/configuration"
	testsuite "github.com/vatriathlon/stopwatch/test/suite"
)

func TestLapRepository(t *testing.T) {
	config, err := configuration.New()
	require.NoError(t, err)
	suite.Run(t, &LapRepositoryTestSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

type LapRepositoryTestSuite struct {
	testsuite.DBTestSuite
	app application.Application
}

func (s *LapRepositoryTestSuite) SetupSuite() {
	s.DBTestSuite.SetupSuite()
	s.app = application.NewGormApplication(s.DB)
}

func (s *LapRepositoryTestSuite) TestCreateLap() {
	ctx := context.Background()
	now := time.Now()
	race := model.Race{
		Name:      fmt.Sprintf("race-%s", uuid.NewV4()),
		StartTime: now,
	}
	err := s.app.Races().Create(ctx, &race)
	require.NoError(s.T(), err)

	team1 := model.Team{
		Name:      "bar1",
		BibNumber: "1",
		RaceID:    race.ID,
	}
	err = s.app.Teams().Create(ctx, &team1)
	require.NoError(s.T(), err)
	team2 := model.Team{
		Name:      "bar2",
		BibNumber: "2",
		RaceID:    race.ID,
	}
	err = s.app.Teams().Create(ctx, &team2)
	require.NoError(s.T(), err)

	s.T().Run("ok", func(t *testing.T) {
		// given
		lap1 := model.Lap{
			RaceID: race.ID,
			TeamID: team1.ID,
			Time:   now.Add(1 * time.Minute),
		}
		// when
		err := s.app.Laps().Create(ctx, &lap1)
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
			err := s.app.Laps().Create(ctx, &lap1)
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
			err := s.app.Laps().Create(ctx, &lap1)
			// then
			require.Error(t, err)
			assert.Equal(t, err.Error(), "missing 'TeamID' field")
		})
	})
}
