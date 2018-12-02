package model

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// Team a team of runner/rider who participates in a given race
type Team struct {
	ID        uuid.UUID `sql:"type:uuid default uuid_generate_v4()" gorm:"primary_key;column:team_id"`
	BibNumber string    `gorm:"column:bib_number"`
	Name      string
	RaceID    uuid.UUID `sql:"type:uuid" gorm:"column:race_id"`
	Laps      []Lap     `gorm:"foreignkey:TeamID"`
}

const (
	teamsTableName = "team"
)

// TableName implements gorm.tabler
func (t Team) TableName() string {
	return teamsTableName
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
	List(raceID uuid.UUID) ([]Team, error)
	FindIDByBibNumber(raceID uuid.UUID, bibnumber string) (uuid.UUID, error)
	LoadByBibNumber(raceID uuid.UUID, bibnumber string) (Team, error)
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
	if team.RaceID == uuid.Nil {
		return errors.New("missing 'RaceID' field")
	}
	if team.Name == "" {
		return errors.New("missing 'Name' field")
	}
	if team.BibNumber == "" {
		return errors.New("missing 'BibNumber' field")
	}
	db := r.db.Create(team)
	if err := db.Error; err != nil {
		return errors.Wrap(err, "fail to store team in DB")
	}
	return nil
}

// List lists all teams for a given race
func (r *GormTeamRepository) List(raceID uuid.UUID) ([]Team, error) {
	result := make([]Team, 0)
	db := r.db.Where("race_id = ?", raceID).Order("bib_number ASC").Find(&result)
	if err := db.Error; err != nil {
		return result, errors.Wrap(err, "fail to list teams")
	}
	return result, nil
}

// FindIDByBibNumber finds the team's ID from the given bibnumber in the given race
func (r *GormTeamRepository) FindIDByBibNumber(raceID uuid.UUID, bibnumber string) (uuid.UUID, error) {
	var team Team
	err := r.db.Raw(
		fmt.Sprintf("select team_id from %s where race_id = ? and bib_number = ?", team.TableName()),
		raceID, bibnumber).Scan(&team).Error
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "fail to find team by bibnumber")
	}
	return team.ID, nil
}

// LoadByBibNumber loads the team along with its laps from the given bibnumber in the given race
func (r *GormTeamRepository) LoadByBibNumber(raceID uuid.UUID, bibnumber string) (Team, error) {
	result := Team{}
	db := r.db.Preload("Laps").Where("race_id = ? and bib_number = ?", raceID, bibnumber).First(&result)
	if err := db.Error; err != nil {
		return result, errors.Wrap(err, "fail to find team by bibnumber")
	}
	return result, nil

}
