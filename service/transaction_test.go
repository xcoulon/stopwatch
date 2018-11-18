package service_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/vatriathlon/stopwatch/configuration"

	"github.com/vatriathlon/stopwatch/service"
	testsuite "github.com/vatriathlon/stopwatch/test/suite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TransactionTestSuite struct {
	testsuite.DBTestSuite
	gormService *service.GormService
}

func TestRunTransaction(t *testing.T) {
	config, err := configuration.New()
	require.NoError(t, err)
	suite.Run(t, &TransactionTestSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

func (s *TransactionTestSuite) SetupTest() {
	s.DBTestSuite.SetupTest()
	s.gormService = service.NewGormService(s.DB)
}

func (s *TransactionTestSuite) TransactionTestSuiteInTime() {
	// given
	computeTime := 10 * time.Second
	// then
	err := service.Transactional(s.gormService, func(r service.Repositories) error {
		time.Sleep(computeTime)
		return nil
	})
	// then
	require.NoError(s.T(), err)
}

func (s *TransactionTestSuite) TransactionTestSuiteOut() {
	// given
	computeTime := 6 * time.Minute
	service.SetDatabaseTransactionTimeout(5 * time.Second)
	// then
	err := service.Transactional(s.gormService, func(r service.Repositories) error {
		time.Sleep(computeTime)
		return nil
	})
	// then
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "database transaction timeout")
}

func (s *TransactionTestSuite) TransactionTestSuitePanicAndRecoverWithStack() {
	// then
	err := service.Transactional(s.gormService, func(r service.Repositories) error {
		bar := func(a, b interface{}) {
			// This comparison while legal at compile time will cause a runtime
			// error like this: "comparing uncomparable type
			// map[string]interface {}". The transaction will panic and recover
			// but you will probably never find out where the error came from if
			// the stack is not captured in the transaction recovery. This test
			// ensures that the stack is captured.
			if a == b {
				fmt.Printf("never executed")
			}
		}
		foo := func() {
			a := map[string]interface{}{}
			b := map[string]interface{}{}
			bar(a, b)
		}
		foo()
		return nil
	})
	// then
	require.Error(s.T(), err)
	// ensure there's a proper stack trace that contains the name of this test
	require.Contains(s.T(), err.Error(), "(*TransactionTestSuite).TransactionTestSuitePanicAndRecoverWithStack.func1(")
}
