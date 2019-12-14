package service

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vatriathlon/stopwatch/pkg/model"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ResultService the interface for the application service
type ResultService struct {
	baseService *GormService
	raceRepo    model.RaceRepository
	teamRepo    model.TeamRepository
	lapRepo     model.LapRepository
}

// NewResultService returns a new ResultService
func NewResultService(db *gorm.DB) ResultService {
	return ResultService{
		baseService: NewGormService(db),
		raceRepo:    model.NewRaceRepository(db),
		teamRepo:    model.NewTeamRepository(db),
		lapRepo:     model.NewLapRepository(db),
	}
}

type teamResult struct {
	bibNumber string
	name      string
	category  string
	gender    string
	members   string
	club      string
	laps      int
	totalTime time.Duration
}

const (
	scratchQuery = `select t.bib_number, t.name, t.gender, t.age_category, t.challenge, 
		member1_last_name, member1_first_name, member1_club, 
		member2_last_name, member2_first_name, member1_club, 
		count(l), max(l.time)
		from team t join lap l on l.team_id = t.team_id 
		where t.race_id = ? 
		group by 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11
		order by 12 desc, 13 asc;`
	entrepriseChallengeQuery = `select t.bib_number, t.name, t.gender, t.age_category, t.challenge, 
		member1_last_name, member1_first_name, member1_club, 
		member2_last_name, member2_first_name, member1_club, 
		count(l), max(l.time)
		from team t join lap l on l.team_id = t.team_id 
		where t.race_id = ? and t.challenge = 'Challenge Entreprise'
		group by 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11
		order by 12 desc, 13 asc;`
	byGenderAndAgeQuery = `select t.bib_number, t.name, t.gender, t.age_category, t.challenge, 
		member1_last_name, member1_first_name, member1_club, 
		member2_last_name, member2_first_name, member1_club, 
		count(l), max(l.time)
		from team t join lap l on l.team_id = t.team_id 
		where t.race_id = ? and t.age_category = ? and t.gender = ?
		group by 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11
		order by 12 desc, 13 asc;`
)

// GenerateResults imports the data from the given file
func (s *ResultService) GenerateResults(raceID int, outputDir string) ([]string, error) {
	files := []string{}
	race, err := s.raceRepo.Lookup(raceID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to generate results")
	}
	// scratch
	scratchRows, err := s.baseService.db.Raw(scratchQuery, race.ID).Rows()
	if err != nil {
		return nil, errors.Wrap(err, "unable to generate results")
	}
	defer scratchRows.Close()
	file, err := generateAsciidoc(outputDir, race, scratchRows, "Scratch", "", true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to generate results")
	}
	if file != "" {
		files = append(files, file)
	}
	// challenge entreprises
	// challengeRows, err := s.baseService.db.Raw(entrepriseChallengeQuery, race.ID).Rows()
	// if err != nil {
	// 	return nil, errors.Wrap(err, "unable to generate results")
	// }
	// defer challengeRows.Close()
	// file, err = generateAsciidoc(outputDir, race, challengeRows, "Challenge Entreprise", "", true)
	// if err != nil {
	// 	return files, errors.Wrap(err, "unable to generate results")
	// }
	// if file != "" {
	// 	files = append(files, file)
	// }

	// by age and gender
	ageCategories := []string{Poussin, Pupille, Benjamin, Minime, Cadet, Junior, Senior, Veteran}
	genders := []string{"H", "F", "M"}
	for _, ageCategory := range ageCategories {
		for _, gender := range genders {
			categoryRows, err := s.baseService.db.Raw(byGenderAndAgeQuery, race.ID, ageCategory, gender).Rows()
			defer categoryRows.Close()
			if err != nil {
				return nil, errors.Wrap(err, "unable to generate results")
			}
			file, err = generateAsciidoc(outputDir, race, categoryRows, ageCategory, gender, false)
			if err != nil {
				return nil, errors.Wrap(err, "unable to generate results")
			}
			if file != "" {
				files = append(files, file)
			}
		}
	}
	return files, nil
}

func generateAsciidoc(outputDir string, race model.Race, rows *sql.Rows, cat1, cat2 string, includeAgeGender bool) (string, error) {
	results, err := readRows(race, rows)
	if err != nil {
		return "", errors.Wrap(err, "unable to generate results")
	}
	var category string
	if cat2 != "" {
		category = fmt.Sprintf("%s-%s", cat1, cat2)
	} else {
		category = cat1
	}
	if len(results) == 0 {
		logrus.WithField("race_name", race.Name).
			WithField("result_category", category).
			Debug("skipping: no result in this category for this race")
		return "", nil
	}
	file, err := os.Create(filepath.Join(outputDir, fmt.Sprintf("%s-%s.adoc", strings.Replace(race.Name, " ", "-", -1), category)))
	if err != nil {
		return "", errors.Wrap(err, "unable to generate results in asciidoc")
	}
	defer file.Close()

	logrus.WithField("race_name", race.Name).
		WithField("result_category", category).
		WithField("teams", len(results)).
		WithField("file", file.Name()).
		Info("generating results...")

	adocWriter := bufio.NewWriter(file)
	adocWriter.WriteString(fmt.Sprintf("= Classement %s\n\n", label(cat1, cat2)))
	adocWriter.WriteString(fmt.Sprintf("== Classement %s\n\n", label(cat1, cat2)))
	// table header
	adocWriter.WriteString("[cols=\"2,5,5,")
	if includeAgeGender {
		adocWriter.WriteString("5,")
	}
	adocWriter.WriteString("8,8,3,4\"]\n")
	adocWriter.WriteString("|===\n")
	adocWriter.WriteString("|# |Dossard ")
	adocWriter.WriteString("|Equipe ")
	if includeAgeGender {
		adocWriter.WriteString("|Catégorie ")
	}
	adocWriter.WriteString("|Coureurs |Club |Tours |Temps Total\n\n")

	// table rows
	for i, r := range results {
		adocWriter.WriteString(fmt.Sprintf("|%d |%s |%s ",
			i+1,
			r.bibNumber,
			r.name))
		if includeAgeGender {
			adocWriter.WriteString(fmt.Sprintf("|%s ",
				r.category))
		}
		adocWriter.WriteString(fmt.Sprintf("|%s |%s |%d |%s \n",
			r.members,
			r.club,
			r.laps,
			fmtDuration(r.totalTime)))
	}
	// close table
	adocWriter.WriteString("|===\n")
	err = adocWriter.Flush()
	if err != nil {
		return "", errors.Wrap(err, "unable to generate results in asciidoc")
	}
	return file.Name(), nil
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func label(cat1, cat2 string) string {
	// "Scratch" and "Challenge Entreprise"
	if cat2 == "" {
		return cat1
	}
	// other: age / gender
	switch cat2 {
	case "M":
		return fmt.Sprintf("%ss / Mixte", cat1)
	case "F":
		return fmt.Sprintf("%ss / Femmes", cat1)
	default:
		return fmt.Sprintf("%ss / Hommes", cat1)
	}
}

func readRows(race model.Race, rows *sql.Rows) ([]teamResult, error) {
	defer rows.Close()
	results := []teamResult{}
	for rows.Next() {
		var bibNumber int
		var name string
		var gender string
		var ageCategory string
		var challenge string
		var member1LastName string
		var member1FirstName string
		var member1Club string
		var member2LastName string
		var member2FirstName string
		var member2Club string
		var laps int
		var endTime time.Time
		err := rows.Scan(&bibNumber, &name, &gender, &ageCategory, &challenge, &member1LastName,
			&member1FirstName, &member1Club, &member2LastName, &member2FirstName, &member1Club,
			&laps, &endTime)
		if err != nil {
			return results, errors.Wrap(err, "unable to generate results")
		}

		result := teamResult{
			bibNumber: strconv.Itoa(bibNumber),
			name:      name,
			category:  getCategory(ageCategory, gender),
			members:   getMemberNames(member1LastName, member2LastName),
			club:      getMemberClubs(member1Club, member2Club),
			laps:      laps,
			totalTime: endTime.Sub(race.StartTime).Round(time.Second),
		}
		logrus.WithField("name", result.name).
			WithField("laps", result.laps).
			WithField("total_time", result.totalTime).
			Debug("adding team to result")
		results = append(results, result)
	}
	return results, nil
}

func getCategory(ageCategory, gender string) string {
	return fmt.Sprintf("%s/%s", string([]rune(ageCategory)[0]), string([]rune(gender)[0]))
}

func getMemberNames(lastName1, lastName2 string) string {
	return fmt.Sprintf("%s - %s", lastName1, lastName2)
}

func getMemberClubs(member1Club, member2Club string) string {
	if member1Club == member2Club {
		return member1Club
	}
	return strings.TrimSpace(fmt.Sprintf("%s %s", member1Club, member2Club))
}
