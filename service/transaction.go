package service

import (
	"time"

	"github.com/fabric8-services/fabric8-common/log"

	"github.com/pkg/errors"
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
		log.Error(nil, map[string]interface{}{
			"err": err,
		}, "database BeginTransaction failed!")

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
				log.Debug(nil, nil, "Rolling back the transaction...")
				tx.Rollback()
				log.Error(nil, map[string]interface{}{
					"err": err,
				}, "database transaction failed!")
				return errors.WithStack(err)
			}

			tx.Commit()
			log.Debug(nil, nil, "Commit the transaction!")
			return nil
		case <-txTimeout:
			log.Debug(nil, nil, "Rolling back the transaction...")
			tx.Rollback()
			return errors.New("database transaction timeout")
		}
	}()
}
