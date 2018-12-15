package service_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/vatriathlon/stopwatch/service"

	"github.com/vatriathlon/stopwatch/model"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/vatriathlon/stopwatch/configuration"
	testmodel "github.com/vatriathlon/stopwatch/test/model"
	testsuite "github.com/vatriathlon/stopwatch/test/suite"
)

func TestResultService(t *testing.T) {
	config, err := configuration.New()
	require.NoError(t, err)
	suite.Run(t, &ResultServiceTestSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

type ResultServiceTestSuite struct {
	testsuite.DBTestSuite
}

func (s *ResultServiceTestSuite) TestListRacesNoResult() {
	// given a dataset
	raceRepo := model.NewRaceRepository(s.DB)
	teamRepo := model.NewTeamRepository(s.DB)
	lapRepo := model.NewLapRepository(s.DB)
	race := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err := raceRepo.Create(&race)
	require.NoError(s.T(), err)
	race.StartTime = time.Now().Add(-60 * time.Minute)
	err = raceRepo.Save(&race)
	require.NoError(s.T(), err)

	// 10 teams, 4 laps each
	for i := 1; i < 120; i++ {
		team := testmodel.NewTeam(race.ID, i)
		// every 10: 'entreprise' challenge
		if i%10 == 0 {
			team.Challenge = "Challenge Entreprise"
		}
		if i%2 == 0 {
			team.Gender = "F"
		}
		if i%3 == 0 {
			team.Gender = "H"
		}
		if i%3 == 0 {
			team.AgeCategory = service.Veteran
		}
		if i%4 == 0 {
			team.AgeCategory = service.Cadet
		}
		if i%5 == 0 {
			team.Gender = "M"
		}
		if i%6 == 0 {
			team.Challenge = "Challenge Entreprise"
		}

		err := teamRepo.Create(&team)
		require.NoError(s.T(), err)
		for j := 0; j < 4; j++ {
			// teams 3 and 4 did not finish
			if j == 3 && (i == 3 || i == 4) {
				continue
			}

			// team 5 did not start
			if i == 5 {
				continue
			}
			lapTime := time.Duration((15*(j+1))+i) * time.Minute // 15min per lap, each time a bit slower (based on bibnumber)
			lap := model.Lap{
				RaceID: race.ID,
				TeamID: team.ID,
				Time:   race.StartTime.Add(lapTime),
			}
			err := lapRepo.Create(&lap)
			require.NoError(s.T(), err)
		}
	}

	svc := service.NewResultService(s.DB)
	// when
	err = svc.GenerateResults(race.ID, "../tmp/results")
	// then
	require.NoError(s.T(), err)

}
