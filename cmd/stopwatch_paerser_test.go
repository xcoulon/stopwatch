package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	"github.com/vatriathlon/stopwatch/cmd"
)

func TestUseRaceCmd(t *testing.T) {
	// given
	actual := `use adults race  `
	// when
	result, err := cmd.Parse("cmd", []byte(actual))
	// then
	require.NoError(t, err)
	require.IsType(t, cmd.UseRaceCmd{}, result)
	c := result.(cmd.UseRaceCmd)
	assert.Equal(t, "adults race", c.RaceName)
}
