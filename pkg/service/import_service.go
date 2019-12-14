package service

import (
	"encoding/csv"
	"io"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/vatriathlon/stopwatch/pkg/model"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
						return errors.Wrapf(err, "unable to create team member from %+v", record)
					}
				} else {
					teamMember2, err = newTeamMember(record)
					if err != nil {
						return errors.Wrapf(err, "unable to create team member from %+v", record)
					}
					var err error
					bibNumber, err := strconv.Atoi(record[1])
					if err != nil {
						return errors.Wrapf(err, "unable to convert bibnumber '%s' to a number", record[1])
					}
					team := model.Team{
						Name:        record[2], // team name
						AgeCategory: GetTeamAgeCategory(teamMember1.AgeCategory, teamMember2.AgeCategory),
						// Challenge:   record[3], // race choice (open/entreprise)
						BibNumber: bibNumber,
						Member1:   teamMember1,
						Member2:   teamMember2,
						Gender:    genderFrom(teamMember1, teamMember2),
						RaceID:    races[record[0]].ID,
					}
					err = app.Teams().Create(&team)
					if err != nil {
						return errors.Wrapf(err, "unable to create team from %+v", team)
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
	dateOfBirth, err := time.Parse("02/01/2006", record[5])
	if err != nil {
		return model.TeamMember{}, errors.Wrapf(err, "unable to parse date '%s'", record[5])
	}
	return model.TeamMember{
		LastName:    record[3],
		FirstName:   record[4],
		DateOfBirth: dateOfBirth,
		Gender:      record[6],
		AgeCategory: GetAgeCategory(dateOfBirth),
		Club:        record[9],
	}, nil
}

const (
	// MiniPoussin 2012 à 2013
	MiniPoussin = "Mini-poussin"
	// Poussin 		2010 à 2011
	Poussin = "Poussin"
	// Pupille 		2008 à 2009
	Pupille = "Pupille"
	// Benjamin 	2006 à 2007
	Benjamin = "Benjamin"
	// Minime 		2004 à 2005
	Minime = "Minime"
	// Cadet 		2002 à 2003
	Cadet = "Cadet"
	// Junior 		2000 à 2001
	Junior = "Junior"
	// Senior 	 	1980 à 1999
	Senior = "Senior"
	// Veteran 		1955 à 1979
	Veteran = "Vétéran"
)

// GetAgeCategory gets the age category associated with the given date of birth
func GetAgeCategory(dateOfBirth time.Time) string {
	yearOfBirth := dateOfBirth.Year()
	logrus.WithField("year_of_birth", yearOfBirth).Debug("computing age category")
	switch {
	case yearOfBirth == 2012 || yearOfBirth == 2013:
		return MiniPoussin
	case yearOfBirth == 2010 || yearOfBirth == 2011:
		return Poussin
	case yearOfBirth == 2008 || yearOfBirth == 2009:
		return Pupille
	case yearOfBirth == 2006 || yearOfBirth == 2007:
		return Benjamin
	case yearOfBirth == 2004 || yearOfBirth == 2005:
		return Minime
	case yearOfBirth == 2002 || yearOfBirth == 2003:
		return Cadet
	case yearOfBirth == 2000 || yearOfBirth == 2001:
		return Junior
	case yearOfBirth >= 1980 && yearOfBirth <= 1999:
		return Senior
	default:
		return Veteran
	}

}

var ageCategories map[string]int

func init() {
	ageCategories = map[string]int{
		MiniPoussin: 1,
		Poussin:     2,
		Pupille:     3,
		Benjamin:    4,
		Minime:      5,
		Cadet:       6,
		Junior:      7,
		Veteran:     8,
		Senior:      9,
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
