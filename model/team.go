package model

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// Team a team of 2 runner/rider who participates in a given race
type Team struct {
	ID        int        `gorm:"primary_key;column:team_id"`
	Name      string     `gorm:"column:name"`
	Gender    string     `gorm:"column:gender"`
	Challenge string     `gorm:"column:challenge"`
	Category  string     `gorm:"column:category"`
	BibNumber string     `gorm:"column:bib_number"`
	Member1   TeamMember `gorm:"embedded;embedded_prefix:member1_"`
	Member2   TeamMember `gorm:"embedded;embedded_prefix:member2_"`
	RaceID    int        `gorm:"column:race_id"`
	Laps      []Lap      `gorm:"foreignkey:TeamID"`
}

// TeamMember a member of a team
type TeamMember struct {
	FirstName   string `gorm:"column:first_name"`
	LastName    string
	DateOfBirth time.Time
	Gender      string
}

const (
	teamTableName = "team"
)

// TableName implements gorm.tabler
func (t Team) TableName() string {
	return teamTableName
}

// Ensure Team implements the Equaler interface
var _ Equaler = Team{}
var _ Equaler = (*Team)(nil)

// Equal returns true if two AuditLog objects are equal; otherwise false is returned.
func (t Team) Equal(o Equaler) bool {
	other, ok := o.(Team)
	if !ok {
		return false
	}
	return t.ID == other.ID
}

// TeamRepository provides functions to create and view teams
type TeamRepository interface {
	Create(team *Team) error
	List(raceID int) ([]Team, error)
	FindIDByBibNumber(raceID int, bibnumber string) (int, error)
	LoadByBibNumber(raceID int, bibnumber string) (Team, error)
}

// NewTeamRepository creates a new GormTeamRepository
func NewTeamRepository(db *gorm.DB) TeamRepository {
	repository := &GormTeamRepository{
		db: db,
	}
	return repository
}

// GormTeamRepository implements Repository using gorm
type GormTeamRepository struct {
	db *gorm.DB
}

// Create stores the given team
func (r *GormTeamRepository) Create(team *Team) error {
	// check values
	if team == nil {
		return errors.New("missing team to persist")
	}
	if team.BibNumber == "" {
		return errors.New("missing 'BibNumber' field")
	}
	if team.RaceID == 0 {
		return errors.New("missing 'RaceID' field")
	}
	if team.Name == "" {
		return errors.New("missing 'Name' field")
	}
	db := r.db.Create(team)
	if err := db.Error; err != nil {
		return errors.Wrap(err, "fail to store team in DB")
	}
	return nil
}

// List lists all teams for a given race
func (r *GormTeamRepository) List(raceID int) ([]Team, error) {
	result := make([]Team, 0)
	db := r.db.Preload("Laps").Where("race_id = ?", raceID).Order("bib_number ASC").Find(&result)
	if err := db.Error; err != nil {
		return result, errors.Wrap(err, "fail to list teams")
	}
	return result, nil
}

// FindIDByBibNumber finds the team's ID from the given bibnumber in the given race
func (r *GormTeamRepository) FindIDByBibNumber(raceID int, bibnumber string) (int, error) {
	var team Team
	err := r.db.Raw(
		fmt.Sprintf("select team_id from %s where race_id = ? and bib_number = ?", team.TableName()),
		raceID, bibnumber).Scan(&team).Error
	if err != nil {
		return -1, errors.Wrapf(err, "fail to find team with bibnumber '%s' in race with id='%d'", bibnumber, raceID)
	}
	return team.ID, nil
}

// LoadByBibNumber loads the team along with its laps from the given bibnumber in the given race
func (r *GormTeamRepository) LoadByBibNumber(raceID int, bibnumber string) (Team, error) {
	result := Team{}
	db := r.db.Preload("Laps").Where("race_id = ? and bib_number = ?", raceID, bibnumber).First(&result)
	if err := db.Error; err != nil {
		return result, errors.Wrap(err, "fail to find team by bibnumber")
	}
	return result, nil

}
