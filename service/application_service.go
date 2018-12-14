package service

import (
	"encoding/csv"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

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

	return Transactional(s.baseService, func(app Repositories) error {
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
					var err error
					bibNumber, err := strconv.Atoi(record[1])
					if err != nil {
						return errors.Wrapf(err, "unable to convert bibnumber '%s' to a number", record[10])
					}
					team := model.Team{
						Name:        record[2], // team name
						AgeCategory: GetTeamAgeCategory(teamMember1.AgeCategory, teamMember2.AgeCategory),
						Challenge:   record[3], // race choice (open/entreprise)
						BibNumber:   bibNumber,
						Member1:     teamMember1,
						Member2:     teamMember2,
						Gender:      genderFrom(teamMember1, teamMember2),
						RaceID:      races[record[0]].ID,
					}
					err = app.Teams().Create(&team)
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
	})
}

func genderFrom(teamMember1, teamMember2 model.TeamMember) string {
	genders := []string{teamMember1.Gender, teamMember2.Gender}
	sort.Strings(genders)
	return strings.Join(genders, "")
}

func newTeamMember(record []string) (model.TeamMember, error) {
	dateOfBirth, err := time.Parse("02/01/2006", record[6])
	if err != nil {
		return model.TeamMember{}, errors.Wrapf(err, "unable to parse date '%s'", record[3])
	}
	return model.TeamMember{
		LastName:    record[4],
		FirstName:   record[5],
		DateOfBirth: dateOfBirth,
		Gender:      record[7],
		AgeCategory: GetAgeCategory(dateOfBirth),
		Club:        record[10],
	}, nil
}

const (
	// Poussin 		2009 à 2013
	Poussin = "Poussin"
	// Pupille 		2007 à 2008
	Pupille = "Pupille"
	// Benjamin 	2005 à 2006
	Benjamin = "Benjamin"
	// Minime 		2003 à 2004
	Minime = "Minime"
	// Cadet 		2001 à 2002
	Cadet = "Cadet"
	// Junior 		1999 à 2000
	Junior = "Junior"
	// Senior 	 	1979 à 1998
	Senior = "Senior"
	// Veteran 		1955 à 1978
	Veteran = "Vétéran"
)

var ageCategories map[string]int

func init() {
	ageCategories = map[string]int{
		Poussin:  1,
		Pupille:  2,
		Benjamin: 3,
		Minime:   4,
		Cadet:    5,
		Junior:   6,
		Veteran:  7,
		Senior:   8,
	}
}

// GetAgeCategory gets the age category associated with the given date of birth
func GetAgeCategory(dateOfBirth time.Time) string {
	yearOfBirth := dateOfBirth.Year()
	logrus.WithField("year_of_birth", yearOfBirth).Debug("computing age category")
	if yearOfBirth >= 2009 {
		return Poussin
	}
	if yearOfBirth == 2007 || yearOfBirth == 2008 {
		return Pupille
	}
	if yearOfBirth == 2005 || yearOfBirth == 2006 {
		return Benjamin
	}
	if yearOfBirth == 2003 || yearOfBirth == 2004 {
		return Minime
	}
	if yearOfBirth == 2001 || yearOfBirth == 2002 {
		return Cadet
	}
	if yearOfBirth == 1999 || yearOfBirth == 2000 {
		return Junior
	}
	if yearOfBirth >= 1979 && yearOfBirth <= 1998 {
		return Senior
	}
	return Veteran
}

// GetTeamAgeCategory computes the age category for the team
func GetTeamAgeCategory(ageCategory1, ageCategory2 string) string {
	teamAgeCategoryValue := math.Max(float64(ageCategories[ageCategory1]), float64(ageCategories[ageCategory2]))
	logrus.WithField("team_age_category_value", teamAgeCategoryValue).Debugf("computing team age category...")
	for k, v := range ageCategories {
		if float64(v) == teamAgeCategoryValue {
			return k
		}
	}
	return ""

}
