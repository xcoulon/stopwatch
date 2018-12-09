package service

import (
	"encoding/csv"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
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
		return app.Races().Start(&race)
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
func (s *ApplicationService) AddLap(raceID int, bibnumber string) (model.Team, error) {
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

// ImportFromFile imports the data from the given file
func (s *ApplicationService) ImportFromFile(filename string) error {
	// list races once for all and map by name
	races := map[string]model.Race{}
	err := Transactional(s.baseService, func(app Repositories) error {
		all, err := app.Races().List()
		if err != nil {
			return err
		}
		for _, r := range all {
			races[r.Name] = r
		}
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "unable to load data")
	}

	var headers []string
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	r := csv.NewReader(file)
	undefinedMember := model.TeamMember{}
	teamMember1 := undefinedMember
	teamMember2 := undefinedMember

	bibNumberSeq := 1
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if headers == nil {
			headers = record
		} else {
			if teamMember1 == undefinedMember {
				teamMember1, err = newTeamMember(record)
				if err != nil {
					return errors.Wrapf(err, "unable to create team member from %v", record)
				}
			} else {
				teamMember2, err = newTeamMember(record)
				if err != nil {
					return errors.Wrapf(err, "unable to create team member from %v", record)
				}
				bibNumber := record[10]
				if bibNumber == "" {
					bibNumberSeq++
					bibNumber = strconv.Itoa(bibNumberSeq)
				}
				team := model.Team{
					Name:      record[9],
					Category:  record[5],
					Challenge: record[11],
					BibNumber: bibNumber,
					Member1:   teamMember1,
					Member2:   teamMember2,
					Gender:    genderFrom(teamMember1, teamMember2),
					RaceID:    races[record[5]].ID,
				}
				err := Transactional(s.baseService, func(app Repositories) error {
					return app.Teams().Create(&team)
				})
				if err != nil {
					return errors.Wrapf(err, "unable to create team from %v", team)
				}
				// reset
				teamMember1 = undefinedMember
				teamMember2 = undefinedMember
			}
		}
	}
	return nil
}

func genderFrom(teamMember1, teamMember2 model.TeamMember) string {
	genders := []string{teamMember1.Gender, teamMember2.Gender}
	sort.Strings(genders)
	return strings.Join(genders, "")
}

func newTeamMember(record []string) (model.TeamMember, error) {
	dateOfBirth, err := time.Parse("02/01/2006", record[3])
	if err != nil {
		return model.TeamMember{}, errors.Wrapf(err, "unable to parse date '%s'", record[3])
	}
	return model.TeamMember{
		FirstName:   record[1],
		LastName:    record[2],
		DateOfBirth: dateOfBirth,
		Gender:      record[4],
	}, nil
}
