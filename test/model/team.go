package testmodel

import (
	"fmt"
	"time"

	"github.com/vatriathlon/stopwatch/model"
)

// NewTeam returns a new team
func NewTeam(raceID int, bibnumber int) model.Team {
	return model.Team{
		Name:        fmt.Sprintf("team %d", bibnumber),
		BibNumber:   bibnumber,
		RaceID:      raceID,
		Gender:      "HF",
		Challenge:   "open",
		AgeCategory: "Senior",
		Member1:     newTeamMember("john", "doe", "H"),
		Member2:     newTeamMember("jane", "doe", "F"),
	}
}

func newTeamMember(firstname, lastname, gender string) model.TeamMember {
	return model.TeamMember{
		FirstName:   firstname,
		LastName:    lastname,
		DateOfBirth: time.Now().Add(-30 * 12 * 24 * time.Hour), // 30 years old
		AgeCategory: "Senior",
		Gender:      gender,
	}
}
