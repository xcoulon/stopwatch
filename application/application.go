package application

import (
	"github.com/vatriathlon/stopwatch/model"
)

//An Application stands for a particular implementation of the business logic of our application
type Application interface {
	Races() model.RaceRepository
	Teams() model.TeamRepository
	Laps() model.LapRepository
}

// A Transaction abstracts a database transaction. The repositories created for the transaction object make changes inside the the transaction
type Transaction interface {
	Application
	Commit() error
	Rollback() error
}

// A DB stands for a particular database (or a mock/fake thereof). It also includes "Application" for creating transactionless repositories
type DB interface {
	Application
	BeginTransaction() (Transaction, error)
}
