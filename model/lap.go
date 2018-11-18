package model

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// Lap a lap for a given team in a given race
type Lap struct {
	ID     uuid.UUID `sql:"type:uuid default uuid_generate_v4()" gorm:"primary_key;column:lap_id"`
	Time   time.Time
	RaceID uuid.UUID `sql:"type:uuid" gorm:"column:race_id"`
	TeamID uuid.UUID `sql:"type:uuid" gorm:"column:team_id"`
}

const (
	lapsTableName = "lap"
)

// TableName implements gorm.tabler
func (l Lap) TableName() string {
	return lapsTableName
}

// Ensure Lap implements the Equaler interface
var _ Equaler = Lap{}
var _ Equaler = (*Lap)(nil)

// Equal returns true if two Lap objects are equal; otherwise false is returned.
func (l Lap) Equal(o Equaler) bool {
	other, ok := o.(Lap)
	if !ok {
		return false
	}
	return l.ID == other.ID
}

// LapRepository provides functions to create and view team laps
type LapRepository interface {
	Create(ctx context.Context, lap *Lap) error
}

// NewLapRepository creates a new GormLapRepository
func NewLapRepository(db *gorm.DB) LapRepository {
	repository := &GormLapRepository{
		db: db,
	}
	return repository
}

// GormLapRepository implements Repository using gorm
type GormLapRepository struct {
	db *gorm.DB
}

// Create stores the given lap
func (r *GormLapRepository) Create(ctx context.Context, lap *Lap) error {
	// check values
	if lap == nil {
		return errors.New("missing lap to create")
	}
	if lap.RaceID == uuid.Nil {
		return errors.New("missing 'RaceID' field")
	}
	if lap.TeamID == uuid.Nil {
		return errors.New("missing 'TeamID' field")
	}
	db := r.db.Create(lap)
	if err := db.Error; err != nil {
		return errors.Wrap(err, "fail to store lap in DB")
	}
	return nil
}
