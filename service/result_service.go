package service

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vatriathlon/stopwatch/model"

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
	name      string
	category  string
	gender    string
	member1   string
	member2   string
	club      string
	laps      int
	totalTime time.Duration
}

// GenerateResults imports the data from the given file
func (s *ResultService) GenerateResults(raceID int, outputDir string) error {
	race, err := s.raceRepo.Lookup(raceID)
	if err != nil {
		return errors.Wrap(err, "unable to generate results")
	}
	// scratch
	results := []teamResult{}
	rows, err := s.baseService.db.Raw(
		`select t.bib_number, t.name, t.gender, t.age_category, t.challenge, 
			member1_last_name, member1_first_name, member1_club, 
			member2_last_name, member2_first_name, member1_club, 
			count(l), max(l.time)
			from team t join lap l on l.team_id = t.team_id 
			where t.race_id = ? 
			group by 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11
			order by 12 desc, 13 asc;`, race.ID).Rows()
	if err != nil {
		return errors.Wrap(err, "unable to generate results")
	}
	defer rows.Close()
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
			return errors.Wrap(err, "unable to generate results")
		}

		result := teamResult{
			name:      name,
			category:  getCategory(ageCategory, gender),
			member1:   getMemberName(member1FirstName, member1LastName),
			member2:   getMemberName(member2FirstName, member2LastName),
			club:      getMemberClubs(member1Club, member2Club),
			laps:      laps,
			totalTime: endTime.Sub(race.StartTime),
		}
		logrus.WithField("name", result.name).WithField("laps", result.laps).WithField("total_time", result.totalTime).Info("adding team to result")
		results = append(results, result)
	}

	return generateCSV(outputDir, race, results)
}

func getCategory(ageCategory, gender string) string {
	switch gender {
	case "HH":
		return fmt.Sprintf("%s %s", ageCategory, "H")
	case "FF":
		return fmt.Sprintf("%s %s", ageCategory, "F")
	case "HF", "FH":
		return fmt.Sprintf("%s %s", ageCategory, "Mixte")
	}
	return ""
}

func getMemberName(firstName, lastName string) string {
	return fmt.Sprintf("%s %s", strings.ToTitle(firstName), strings.ToTitle(lastName))
}

func getMemberClubs(member1Club, member2Club string) string {
	return strings.TrimSpace(fmt.Sprintf("%s %s", member1Club, member2Club))
}

func generateCSV(outputDir string, race model.Race, results []teamResult) error {
	file, err := os.Create(filepath.Join(outputDir, fmt.Sprintf("%s.csv", race.Name)))
	if err != nil {
		return errors.Wrap(err, "unable to generate csv")
	}
	defer file.Close()
	csvWriter := csv.NewWriter(file)
	defer csvWriter.Flush()
	// headers
	err = csvWriter.Write([]string{
		"Equipe",
		"Catégorie",
		"Coureur 1",
		"Coureur 2",
		"Club",
		"Nb Tours",
		"Temps Total",
	})
	if err != nil {
		return errors.Wrap(err, "unable to generate csv")
	}

	for _, r := range results {
		err := csvWriter.Write([]string{
			r.name,
			r.category,
			r.member1,
			r.member2,
			r.club,
			strconv.Itoa(r.laps),
			r.totalTime.String(),
		})
		if err != nil {
			return errors.Wrap(err, "unable to generate csv")
		}
	}
	return nil
}

// const (
// 	// FontName the name of the font to use
// 	fontName   string = "SourceCodePro-Regular"
// 	leftMargin string = "  "
// )

// func generatePDF(outputFilename string, race model.Race, results []teamResult) error {
// 	pdf := gopdf.GoPdf{}
// 	pdf.Start(gopdf.Config{PageSize: gopdf.PageSizeA4})
// 	fontLocation := fmt.Sprintf("../ttf/%s.ttf", fontName)
// 	var parser core.TTFParser
// 	err := parser.Parse(fontLocation)
// 	if err != nil {
// 		return errors.Wrap(err, "unable to generate PDF")
// 	}
// 	err = pdf.AddTTFFont(fontName, fontLocation)
// 	if err != nil {
// 		return errors.Wrap(err, "unable to generate PDF")
// 	}

// 	pdf.AddPage()
// 	// title
// 	fontSize := 20
// 	err = pdf.SetFont(fontName, "", fontSize)
// 	if err != nil {
// 		return errors.Wrap(err, "unable to generate PDF")
// 	}
// 	pdf.Br(getHeight(&parser, fontSize) * 2)
// 	pdf.Cell(nil, race.Name)
// 	pdf.Br(getHeight(&parser, fontSize) * 2)
// 	// teams in order
// 	for i, teamResult := range results {
// 		if (i+1)%25 == 0 {
// 			pdf.AddPage()
// 			pdf.Br(getHeight(&parser, fontSize) * 2)
// 		}
// 		// userName := strings.ToTitle(userData[0])
// 		// pdf.Cell(nil, fmt.Sprintf("%s%s", leftMargin, userName))
// 		// pdf.Br(getHeight(&parser, fontSize) * 2)
// 		// items
// 		fontSize = 10
// 		err = pdf.SetFont(fontName, "", fontSize)
// 		if err != nil {
// 			return errors.Wrap(err, "unable to generate PDF")
// 		}
// 		err = pdf.Cell(nil, teamResult.name)
// 		if err != nil {
// 			return errors.Wrap(err, "unable to generate PDF")
// 		}
// 		pdf.Br(getHeight(&parser, fontSize) * 2)

// 	}

// 	// write output
// 	err = pdf.WritePdf(outputFilename)
// 	if err != nil {
// 		return errors.Wrap(err, "unable to generate PDF")
// 	}
// 	return nil
// }

// func getHeight(parser *core.TTFParser, fontSize int) float64 {
// 	//Measure Height
// 	//get  CapHeight (https://en.wikipedia.org/wiki/Cap_height)
// 	cap := float64(float64(parser.CapHeight()) * 1000.00 / float64(parser.UnitsPerEm()))
// 	//convert
// 	realHeight := cap * (float64(fontSize) / 1000.0)
// 	// fmt.Printf("realHeight = %f", realHeight)
// 	return realHeight * 2
// }
