package service

import (
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var databaseTransactionTimeout = 1 * time.Minute

func SetDatabaseTransactionTimeout(t time.Duration) {
	databaseTransactionTimeout = t
}

func DatabaseTransactionTimeout() time.Duration {
	return databaseTransactionTimeout
}

// TransactionManager manages the lifecycle of a database transaction. The transactional resources (such as repositories)
// created for the transaction object make changes inside the transaction
type TransactionManager interface {
	BeginTransaction() (Transaction, error)
}

// Transactional executes the given function in a transaction. If todo returns an error, the transaction is rolled back
func Transactional(svc *GormService, todo func(r Repositories) error) error {
	var tx Transaction
	var err error
	if tx, err = svc.BeginTransaction(); err != nil {
		logrus.WithError(err).Error("database BeginTransaction failed!")
		return errors.WithStack(err)
	}

	return func() error {
		errorChan := make(chan error, 1)
		txTimeout := time.After(databaseTransactionTimeout)

		go func(r Repositories) {
			defer func() {
				if err := recover(); err != nil {
					errorChan <- errors.Errorf("unknown error: %v", err)
				}
			}()
			errorChan <- todo(tx)
		}(tx)

		select {
		case err := <-errorChan:
			if err != nil {
				logrus.Warn("Rolling back the transaction...")
				logrus.WithError(err).Error("database transaction failed. Rolling back...")
				if err2 := tx.Rollback(); err2 != nil {
					logrus.WithError(err2).Error("database transaction rollback failed!")
				}
				return errors.WithStack(err)
			}
			if err := tx.Commit(); err != nil {
				logrus.WithError(err).Error("database transaction commit failed!")
			}
			logrus.Debug("Commit the transaction!")
			return nil
		case <-txTimeout:
			logrus.Debug("Rolling back the transaction...")
			logrus.WithError(err).Error("database transaction timeout. Rolling back...")
			if err2 := tx.Rollback(); err2 != nil {
				logrus.WithError(err2).Error("database transaction rollback failed!")
			}
			return errors.New("database transaction timeout")
		}
	}()
}
