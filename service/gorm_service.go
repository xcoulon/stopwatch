package service

import (
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/vatriathlon/stopwatch/model"
)

// A TXIsoLevel specifies the characteristics of the transaction
// See https://www.postgresql.org/docs/9.3/static/sql-set-transaction.html
type TXIsoLevel int8

const (
	// TXIsoLevelDefault doesn't specify any transaction isolation level, instead the connection
	// based setting will be used.
	TXIsoLevelDefault TXIsoLevel = iota

	// TXIsoLevelReadCommitted means "A statement can only see rows committed before it began. This is the default."
	TXIsoLevelReadCommitted

	// TXIsoLevelRepeatableRead means "All statements of the current transaction can only see rows committed before the
	// first query or data-modification statement was executed in this transaction."
	TXIsoLevelRepeatableRead

	// TXIsoLevelSerializable means "All statements of the current transaction can only see rows committed
	// before the first query or data-modification statement was executed in this transaction.
	// If a pattern of reads and writes among concurrent serializable transactions would create a
	// situation which could not have occurred for any serial (one-at-a-time) execution of those
	// transactions, one of them will be rolled back with a serialization_failure error."
	TXIsoLevelSerializable
)

var _ Service = &GormService{}

var _ TransactionManager = &GormService{}

// NewGormService returns a new GormService object that supports transactions
func NewGormService(db *gorm.DB) *GormService {
	return &GormService{db: db, txIsoLevel: ""}
}

type Service interface {
}

type GormService struct {
	TransactionManager
	txIsoLevel string
	db         *gorm.DB
}

// Transaction represents an existing transaction.  It provides access to transactional resources, plus methods to commit or roll back the transaction
type Transaction interface {
	Repositories
	Commit() error
	Rollback() error
}

// Repositories the repositories accessor
type Repositories interface {
	Races() model.RaceRepository
	Teams() model.TeamRepository
	Laps() model.LapRepository
}

type GormTransaction struct {
	GormRepositories
}

func (g *GormService) Repositories() Repositories {
	return &GormRepositories{db: g.db}
}

// SetTransactionIsolationLevel sets the isolation level for
// See also https://www.postgresql.org/docs/9.3/static/sql-set-transaction.html
func (g *GormService) SetTransactionIsolationLevel(level TXIsoLevel) error {
	switch level {
	case TXIsoLevelReadCommitted:
		g.txIsoLevel = "READ COMMITTED"
	case TXIsoLevelRepeatableRead:
		g.txIsoLevel = "REPEATABLE READ"
	case TXIsoLevelSerializable:
		g.txIsoLevel = "SERIALIZABLE"
	case TXIsoLevelDefault:
		g.txIsoLevel = ""
	default:
		return fmt.Errorf("Unknown transaction isolation level: " + strconv.FormatInt(int64(level), 10))
	}
	return nil
}

// BeginTransaction implements TransactionSupport
func (g *GormService) BeginTransaction() (Transaction, error) {
	tx := g.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	if len(g.txIsoLevel) != 0 {
		tx := tx.Exec(fmt.Sprintf("set transaction isolation level %s", g.txIsoLevel))
		if tx.Error != nil {
			return nil, tx.Error
		}
		return &GormTransaction{GormRepositories{tx}}, nil
	}
	return &GormTransaction{GormRepositories{tx}}, nil
}

// GormRepositories is a base struct for gorm implementations of db & transaction
type GormRepositories struct {
	db *gorm.DB
}

func (g *GormRepositories) Races() model.RaceRepository {
	return model.NewRaceRepository(g.db)
}

func (g *GormRepositories) Teams() model.TeamRepository {
	return model.NewTeamRepository(g.db)
}

func (g *GormRepositories) Laps() model.LapRepository {
	return model.NewLapRepository(g.db)
}

func (g *GormRepositories) DB() *gorm.DB {
	return g.db
}

// Commit implements TransactionSupport
func (g *GormTransaction) Commit() error {
	err := g.db.Commit().Error
	g.db = nil
	return errors.WithStack(err)
}

// Rollback implements TransactionSupport
func (g *GormTransaction) Rollback() error {
	err := g.db.Rollback().Error
	g.db = nil
	return errors.WithStack(err)
}
