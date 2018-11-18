package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/vatriathlon/stopwatch/configuration"
	"github.com/vatriathlon/stopwatch/model"
	"github.com/vatriathlon/stopwatch/service"
	testsuite "github.com/vatriathlon/stopwatch/test/suite"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestService(t *testing.T) {
	config, err := configuration.New()
	require.NoError(t, err)
	suite.Run(t, &ServiceTestSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

type ServiceTestSuite struct {
	testsuite.DBTestSuite
}

func (s *ServiceTestSuite) TestListRacesNoResult() {
	// given
	ctx := context.Background()
	raceRepo := model.NewRaceRepository(s.DB)
	race1 := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err := raceRepo.Create(ctx, &race1)
	require.NoError(s.T(), err)
	svc := service.NewApplicationService(s.DB)
	// when
	races, err := svc.ListRaces(context.Background())
	// then
	require.NoError(s.T(), err)
	assert.Len(s.T(), races, 1)
}

func (s *ServiceTestSuite) TestListRacesMultipleResults() {
	// given
	ctx := context.Background()
	raceRepo := model.NewRaceRepository(s.DB)
	race1 := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err := raceRepo.Create(ctx, &race1)
	require.NoError(s.T(), err)
	race2 := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err = raceRepo.Create(ctx, &race2)
	require.NoError(s.T(), err)
	svc := service.NewApplicationService(s.DB)
	// when
	races, err := svc.ListRaces(context.Background())
	// then
	require.NoError(s.T(), err)
	assert.Len(s.T(), races, 2)
}

func (s *ServiceTestSuite) TestUseRace() {
	// given
	ctx := context.Background()
	raceRepo := model.NewRaceRepository(s.DB)
	race1 := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err := raceRepo.Create(ctx, &race1)
	require.NoError(s.T(), err)
	race2 := model.Race{
		Name: fmt.Sprintf("race %s", uuid.NewV4()),
	}
	err = raceRepo.Create(ctx, &race2)
	require.NoError(s.T(), err)
	svc := service.NewApplicationService(s.DB)
	// when
	raceInUse1, err := svc.UseRace(ctx, race1.Name)
	// then
	require.NoError(s.T(), err)
	assert.Equal(s.T(), race1, raceInUse1)
	// check race in use from the service method as well
	raceInUse2, err := svc.CurrentRace()
	assert.Equal(s.T(), race1.Name, raceInUse2)
}
