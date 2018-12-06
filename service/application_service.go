package service

import (
	"time"

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
func (s *ApplicationService) ListRaces() ([]model.Race, error) {
	var result []model.Race
	err := Transactional(s.baseService, func(app Repositories) error {
		var err error
		result, err = app.Races().List()
		return err
	})
	if err != nil {
		return result, errors.Wrapf(err, "unable to list races")
	}
	return result, nil
}

// UseRace set the current race to the one matching the given name
func (s *ApplicationService) UseRace(name string) (model.Race, error) {
	var result model.Race
	err := Transactional(s.baseService, func(app Repositories) error {
		var err error
		result, err = app.Races().FindByName(name)
		return err
	})
	if err != nil { // also covers the case where no race matched the given name
		return result, errors.Wrapf(err, "unable to find race named '%s'", name)
	}
	s.currentRace = result
	return result, nil
}

// RaceInUse returns the race in user, or an error if none is in use yet
func (s *ApplicationService) RaceInUse() (model.Race, error) {
	if s.currentRace == model.UndefinedRace {
		return model.UndefinedRace, errors.New("no race is in use")
	}
	return s.currentRace, nil
}

// StartCurrentRace set the current race to the one matching the given name
func (s *ApplicationService) StartCurrentRace() (time.Time, error) {
	if s.currentRace == model.UndefinedRace {
		return time.Now(), errors.New("no race in use")
	}
	if s.currentRace.IsStarted() {
		return time.Now(), errors.Errorf("current race already started at %v", s.currentRace.StartTimeStr())
	}
	err := Transactional(s.baseService, func(app Repositories) error {
		// TODO: check that no other race has started but not ended yet
		return app.Races().Start(&s.currentRace)
	})
	if err != nil {
		return time.Now(), errors.Wrap(err, "unable to start race")
	}
	return s.currentRace.StartTime, nil
}

// ListTeams list the teams for the current race.
func (s *ApplicationService) ListTeams() ([]model.Team, error) {
	if s.currentRace == model.UndefinedRace {
		return []model.Team{}, errors.New("no race in use")
	}
	var result []model.Team
	err := Transactional(s.baseService, func(app Repositories) error {
		var err error
		result, err = app.Teams().List(s.currentRace.ID)
		return err
	})
	if err != nil {
		return result, errors.Wrapf(err, "unable to list teams in race")
	}
	return result, nil
}

// AddLap record a new lap at the current time for the teams with given bib numbers
func (s *ApplicationService) AddLap(bibnumber string) (model.Team, error) {
	if s.currentRace == model.UndefinedRace {
		return model.Team{}, errors.New("no race in use")
	}
	var team model.Team
	err := Transactional(s.baseService, func(app Repositories) error {
		teamID, err := app.Teams().FindIDByBibNumber(s.currentRace.ID, bibnumber)
		if err != nil {
			return err
		}
		err = app.Laps().Create(&model.Lap{
			RaceID: s.currentRace.ID,
			TeamID: teamID,
			Time:   time.Now(),
		})
		if err != nil {
			return err
		}
		team, err = app.Teams().LoadByBibNumber(s.currentRace.ID, bibnumber)
		return err
	})
	if err != nil {
		return team, errors.Wrapf(err, "unable to add laps to team")
	}

	return team, nil
}
