package service_test

import (
	"testing"
	"time"

	"github.com/vatriathlon/stopwatch/pkg/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAgeCategory(t *testing.T) {

	pattern := "2006-01-02"

	testcases := map[string]struct {
		dateOfBirth string
		expected    string
	}{
		service.Poussin: {
			dateOfBirth: "2009-02-03",
			expected:    service.Poussin,
		},
		service.Pupille: {
			dateOfBirth: "2008-02-03",
			expected:    service.Pupille,
		},
		service.Benjamin: {
			dateOfBirth: "2006-02-03",
			expected:    service.Benjamin,
		},
		service.Cadet: {
			dateOfBirth: "2001-02-03",
			expected:    service.Cadet,
		},
		service.Junior: {
			dateOfBirth: "1999-02-03",
			expected:    service.Junior,
		},
		service.Senior: {
			dateOfBirth: "1980-02-03",
			expected:    service.Senior,
		},
		service.Veteran: {
			dateOfBirth: "1974-02-03",
			expected:    service.Veteran,
		},
	}

	for testname, testdata := range testcases {
		t.Run(testname, func(t *testing.T) {
			// given
			dateOfBirth, err := time.Parse(pattern, testdata.dateOfBirth)
			require.NoError(t, err)
			// when
			result := service.GetAgeCategory(dateOfBirth)
			// then
			assert.Equal(t, testdata.expected, result)
		})
	}
}

func TestGetTeamAgeCategory(t *testing.T) {

	testcases := map[string]struct {
		category1 string
		category2 string
		expected  string
	}{
		"Poussin/Poussin": {
			service.Poussin,
			service.Poussin,
			service.Poussin,
		},
		"Poussin/Pupille": {
			service.Poussin,
			service.Pupille,
			service.Pupille,
		},
		"Benjamin/Minime": {
			service.Benjamin,
			service.Minime,
			service.Minime,
		},
		"Senior/Senior": {
			service.Senior,
			service.Senior,
			service.Senior,
		},
		"Veteran/Senior": {
			service.Veteran,
			service.Senior,
			service.Senior,
		},
	}
	for testname, testdata := range testcases {
		t.Run(testname, func(t *testing.T) {
			// when
			result := service.GetTeamAgeCategory(testdata.category1, testdata.category2)
			// then
			assert.Equal(t, testdata.expected, result)
		})
	}

}
