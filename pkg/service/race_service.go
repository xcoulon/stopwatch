package service

import (
	"time"

	"github.com/vatriathlon/stopwatch/pkg/model"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// RaceService the interface for the application service
type RaceService struct {
	baseService *GormService
}

// NewRaceService returns a new RaceService
func NewRaceService(db *gorm.DB) RaceService {
	return RaceService{
		baseService: NewGormService(db),
	}
}

// ListRaces list the races.
func (s RaceService) ListRaces() ([]model.Race, error) {
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
func (s RaceService) GetRace(id int) (model.Race, error) {
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
func (s RaceService) StartRace(raceID int) (model.Race, error) {
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
		return race, errors.Wrap(err, "unable to start race")
	}
	return race, nil
}

// TODO: is it usefull? (+needs to be updated to the service code)
// End marks the given race as ended (now)
// func (r *GormRaceRepository) End(race *Race) error {
// 	// check values
// 	if !race.IsStarted() {
// 		return errors.New("race has not started yet")
// 	}
// 	if race.IsEnded() {
// 		return errors.Errorf("race already ended at %v", race.EndTime.Format(raceTimeFmt))
// 	}
// 	race.EndTime = time.Now()
// 	db := r.db.Save(race)
// 	if err := db.Error; err != nil {
// 		return errors.Wrap(err, "fail to save race in DB")
// 	}
// 	return nil
// }

// ListTeams list the teams for the current race.
func (s RaceService) ListTeams(raceID int) ([]model.Team, error) {
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
func (s RaceService) AddLap(raceID int, bibnumber int) (model.Team, error) {
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
