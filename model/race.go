package model

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// Race a race
type Race struct {
	ID        uuid.UUID `sql:"type:uuid default uuid_generate_v4()" gorm:"primary_key;column:race_id"`
	Name      string
	StartTime time.Time
}

const (
	racesTableName = "race"
)

// TableName implements gorm.tabler
func (r Race) TableName() string {
	return racesTableName
}

// Ensure Race implements the Equaler interface
var _ Equaler = Race{}
var _ Equaler = (*Race)(nil)

// Equal returns true if two AuditLog objects are equal; otherwise false is returned.
func (r Race) Equal(o Equaler) bool {
	other, ok := o.(Race)
	if !ok {
		return false
	}
	return r.ID == other.ID
}

// RaceRepository provides functions to create and view races
type RaceRepository interface {
	Create(ctx context.Context, race *Race) error
	// ListByTeamID(ctx context.Context, identityID uuid.UUID, start int, limit int) ([]Race, error)
}

// NewRaceRepository creates a new GormRaceRepository
func NewRaceRepository(db *gorm.DB) RaceRepository {
	repository := &GormRaceRepository{
		db: db,
	}
	return repository
}

// GormRaceRepository implements Repository using gorm
type GormRaceRepository struct {
	db *gorm.DB
}

// Create stores the given race
func (r *GormRaceRepository) Create(ctx context.Context, race *Race) error {
	// check values
	if race == nil {
		return errors.New("missing race to persist")
	}
	if race.Name == "" {
		return errors.New("missing 'Name' field")
	}
	db := r.db.Create(race)
	if err := db.Error; err != nil {
		return errors.Wrap(err, "fail to store race in DB")
	}
	return nil
}
