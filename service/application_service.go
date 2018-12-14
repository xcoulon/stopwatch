package service

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/vatriathlon/stopwatch/model"
)

// ApplicationService the interface for the application service
type ApplicationService struct {
	baseService *GormService
}

// NewApplicationService returns a new ApplicationService
func NewApplicationService(db *gorm.DB) ApplicationService {
	return ApplicationService{
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

// GetRace get the race given its ID
func (s *ApplicationService) GetRace(id int) (model.Race, error) {
	var result model.Race
	err := Transactional(s.baseService, func(app Repositories) error {
		var err error
		result, err = app.Races().Lookup(id)
		return err
	})
	if err != nil {
		return result, errors.Wrapf(err, "unable to get race with id=%d", id)
	}
	return result, nil
}

// StartRace set the current race to the one matching the given name
func (s *ApplicationService) StartRace(raceID int) (time.Time, error) {
	var race model.Race
	err := Transactional(s.baseService, func(app Repositories) error {
		var err error
		race, err = app.Races().Lookup(raceID)
		if err != nil {
			return err
		}
		if race.IsStarted() {
			return errors.Errorf("current race already started at %v", race.StartTimeStr())
		}
		race.StartTime = time.Now()
		return app.Races().Save(&race)
	})
	if err != nil {
		return time.Now(), errors.Wrap(err, "unable to start race")
	}
	return race.StartTime, nil
}

// ListTeams list the teams for the current race.
func (s *ApplicationService) ListTeams(raceID int) ([]model.Team, error) {
	var result []model.Team
	err := Transactional(s.baseService, func(app Repositories) error {
		var err error
		result, err = app.Teams().List(raceID)
		return err
	})
	if err != nil {
		return result, errors.Wrapf(err, "unable to list teams in race")
	}
	return result, nil
}

// AddLap record a new lap at the current time for the teams with given bib numbers
func (s *ApplicationService) AddLap(raceID int, bibnumber int) (model.Team, error) {
	var team model.Team
	err := Transactional(s.baseService, func(app Repositories) error {
		teamID, err := app.Teams().FindIDByBibNumber(raceID, bibnumber)
		if err != nil {
			return err
		}
		err = app.Laps().Create(&model.Lap{
			RaceID: raceID,
			TeamID: teamID,
			Time:   time.Now(),
		})
		if err != nil {
			return err
		}
		team, err = app.Teams().LoadByBibNumber(raceID, bibnumber)
		return err
	})
	if err != nil {
		return team, errors.Wrapf(err, "unable to add laps to team")
	}

	return team, nil
}
