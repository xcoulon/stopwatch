package model

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// Lap a lap for a given team in a given race
type Lap struct {
	ID     int `gorm:"primary_key;column:lap_id"`
	Time   time.Time
	RaceID int `gorm:"column:race_id"`
	TeamID int `gorm:"column:team_id"`
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
	Create(lap *Lap) error
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
func (r *GormLapRepository) Create(lap *Lap) error {
	// check values
	if lap == nil {
		return errors.New("missing lap to create")
	}
	if lap.RaceID == 0 {
		return errors.New("missing 'RaceID' field")
	}
	if lap.TeamID == 0 {
		return errors.New("missing 'TeamID' field")
	}
	db := r.db.Create(lap)
	if err := db.Error; err != nil {
		return errors.Wrap(err, "fail to store lap in DB")
	}
	return nil
}
