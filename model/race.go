package model

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

// Race a race
type Race struct {
	ID        uuid.UUID `sql:"type:uuid default uuid_generate_v4()" gorm:"primary_key;column:race_id"`
	Name      string
	startTime time.Time
	endTime   time.Time
}

// the "undefined" race
var UndefinedRace = Race{}

const (
	racesTableName = "race"
)

// TableName implements gorm.tabler
func (r *Race) TableName() string {
	return racesTableName
}

// IsStarted returns 'true' if the race has already started, false otherwise.
func (r *Race) IsStarted() bool {
	log.Debugf("race started at: %v (%t)", r.startTime, r.startTime.IsZero())
	return !r.startTime.IsZero()
}

// StartTime returns the start time of the race, or empty string if it has not started yet
func (r *Race) StartTime() string {
	if r.startTime.IsZero() {
		return ""
	}
	return r.startTime.Format("2006-01-02 15:04:05")
}

// IsEnded returns 'true' if the race has already ended, false otherwise.
func (r *Race) IsEnded() bool {
	return !r.endTime.IsZero()
}

// EndTime returns the end time of the race, or empty string if it has not ended yet
func (r *Race) EndTime() string {
	if r.endTime.IsZero() {
		return ""
	}
	return r.endTime.Format("2006-01-02 15:04:05")
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
	FindByName(ctx context.Context, name string) (Race, error)
	Start(ctx context.Context, race *Race) error
	End(ctx context.Context, race *Race) error
	List(ctx context.Context) ([]Race, error)
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

// Create creates a race
func (r *GormRaceRepository) Create(ctx context.Context, race *Race) error {
	// check values
	if race == nil {
		return errors.New("missing race to create")
	}
	if race.IsStarted() {
		return errors.New("race to create cannot be started yet")
	}
	if race.IsEnded() {
		return errors.New("race to create cannot be ended yet")
	}
	db := r.db.Create(race)
	if err := db.Error; err != nil {
		return errors.Wrap(err, "fail to store race in DB")
	}
	return nil
}

// FindByName find the race with the given name. Returns an error if none was found
func (r *GormRaceRepository) FindByName(ctx context.Context, name string) (Race, error) {
	var result Race
	db := r.db.First(&result, "name = ?", name)
	if err := db.Error; err != nil {
		return result, err
	}
	return result, nil
}

// Start marks the given race as started (now)
func (r *GormRaceRepository) Start(ctx context.Context, race *Race) error {
	// check values
	if race.IsStarted() {
		return errors.Errorf("race already started at %v", race.StartTime())
	}
	race.startTime = time.Now()
	db := r.db.Save(race)
	if err := db.Error; err != nil {
		return errors.Wrap(err, "fail to save race in DB")
	}
	return nil
}

// End marks the given race as ended (now)
func (r *GormRaceRepository) End(ctx context.Context, race *Race) error {
	// check values
	if !race.IsStarted() {
		return errors.New("race has not started yet")
	}
	if race.IsEnded() {
		return errors.Errorf("race already ended at %v", race.EndTime())
	}
	race.endTime = time.Now()
	db := r.db.Save(race)
	if err := db.Error; err != nil {
		return errors.Wrap(err, "fail to save race in DB")
	}
	return nil
}

// List lists all races
func (r *GormRaceRepository) List(ctx context.Context) ([]Race, error) {
	result := make([]Race, 0)
	db := r.db.Order("name ASC").Find(&result)
	if err := db.Error; err != nil {
		return result, errors.Wrap(err, "fail to list races")
	}
	return result, nil
}
