package service

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/vatriathlon/stopwatch/model"
)

// ApplicationService the interface for the application service
type ApplicationService struct {
	currentRace model.Race
	baseService *GormService
}

// NewApplicationService returns a new ApplicationService
func NewApplicationService(db *gorm.DB) *ApplicationService {
	return &ApplicationService{
		baseService: NewGormService(db),
	}
}

// ListRaces list the races.
func (s *ApplicationService) ListRaces(ctx context.Context) ([]model.Race, error) {
	var result []model.Race
	err := Transactional(s.baseService, func(app Repositories) error {
		var err error
		result, err = app.Races().List(ctx)
		return err
	})
	if err != nil {
		return result, errors.Wrapf(err, "unable to list races")
	}
	return result, nil
}

// UseRace set the current race to the one matching the given name
func (s *ApplicationService) UseRace(ctx context.Context, name string) (model.Race, error) {
	var result model.Race
	err := Transactional(s.baseService, func(app Repositories) error {
		var err error
		result, err = app.Races().FindByName(ctx, name)
		return err
	})
	if err != nil { // also covers the case where no race matched the given name
		return result, errors.Wrapf(err, "unable to find race named '%s'", name)
	}
	s.currentRace = result
	return result, nil
}

// StartCurrentRace set the current race to the one matching the given name
func (s *ApplicationService) StartCurrentRace(ctx context.Context) error {
	if s.currentRace == model.UndefinedRace {
		return errors.New("no race in use")
	}
	if s.currentRace.IsStarted() {
		return errors.Errorf("current race already started at %v", s.currentRace.StartTimeStr())
	}
	err := Transactional(s.baseService, func(app Repositories) error {
		// TODO: check that no other race has started but not ended yet
		return app.Races().Start(ctx, &s.currentRace)
	})
	if err != nil {
		return errors.Wrap(err, "unable to start race")
	}
	return nil
}

// ListTeams list the teams for the current race.
func (s *ApplicationService) ListTeams(ctx context.Context) ([]model.Team, error) {
	if s.currentRace == model.UndefinedRace {
		return []model.Team{}, errors.New("no race in use")
	}
	var result []model.Team
	err := Transactional(s.baseService, func(app Repositories) error {
		var err error
		result, err = app.Teams().List(ctx, s.currentRace.ID)
		return err
	})
	if err != nil {
		return result, errors.Wrapf(err, "unable to list teams in race")
	}
	return result, nil
}

// AddLap record a new lap at the current time for all teams with given bib numbers
func (s *ApplicationService) AddLap(bibnumbers ...string) ([]model.Team, error) {
	if s.currentRace == model.UndefinedRace {
		return []model.Team{}, errors.New("no race in use")
	}

	return []model.Team{}, nil
}
