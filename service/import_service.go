package service

import (
	"encoding/csv"
	"io"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vatriathlon/stopwatch/model"
)

// ImportService the interface for the application service
type ImportService struct {
	baseService *GormService
}

// NewImportService returns a new ImportService
func NewImportService(db *gorm.DB) ImportService {
	return ImportService{
		baseService: NewGormService(db),
	}
}

// ImportFromFile imports the data from the given file
func (s *ImportService) ImportFromFile(filename string) error {
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
	if teamMember1.Gender == teamMember2.Gender {
		return teamMember1.Gender
	}
	return "M"
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

// GetTeamAgeCategory computes the age category for the team
func GetTeamAgeCategory(ageCategory1, ageCategory2 string) string {
	cat1 := ageCategories[ageCategory1]
	cat2 := ageCategories[ageCategory2]
	// assign to senior if 1 veteran + 1 under senior
	if (ageCategory1 == Veteran && cat2 <= ageCategories[Junior]) || (ageCategory2 == Veteran && cat1 <= ageCategories[Junior]) {
		return Senior
	}
	teamAgeCategoryValue := math.Max(float64(cat1), float64(cat2))
	logrus.WithField("team_age_category_value", teamAgeCategoryValue).Debugf("computing team age category...")
	//

	for k, v := range ageCategories {
		if float64(v) == teamAgeCategoryValue {
			return k
		}
	}
	return ""

}
