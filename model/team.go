package model

import (
	"context"

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
	Create(ctx context.Context, team *Team) error
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
func (r *GormTeamRepository) Create(ctx context.Context, team *Team) error {
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
	db := r.db.Create(team)
	if err := db.Error; err != nil {
		return errors.Wrap(err, "fail to store team in DB")
	}
	return nil
}
