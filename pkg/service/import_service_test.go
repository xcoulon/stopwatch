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
		service.MiniPoussin: {
			dateOfBirth: "2012-02-03",
			expected:    service.MiniPoussin,
		},
		service.Poussin: {
			dateOfBirth: "2010-02-03",
			expected:    service.Poussin,
		},
		service.Pupille: {
			dateOfBirth: "2009-02-03",
			expected:    service.Pupille,
		},
		service.Benjamin: {
			dateOfBirth: "2007-02-03",
			expected:    service.Benjamin,
		},
		service.Cadet: {
			dateOfBirth: "2002-02-03",
			expected:    service.Cadet,
		},
		service.Junior: {
			dateOfBirth: "2000-02-03",
			expected:    service.Junior,
		},
		service.Senior: {
			dateOfBirth: "1981-02-03",
			expected:    service.Senior,
		},
		service.Veteran: {
			dateOfBirth: "1975-02-03",
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
		"MiniPoussin/MiniPoussin": {
			service.MiniPoussin,
			service.MiniPoussin,
			service.MiniPoussin,
		},
		"MiniPoussin/Poussin": {
			service.MiniPoussin,
			service.Poussin,
			service.Poussin,
		},
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
		"Veteran/Veteran": {
			service.Veteran,
			service.Veteran,
			service.Veteran,
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
