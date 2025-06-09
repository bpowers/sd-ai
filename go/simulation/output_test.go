package simulation

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/UB-IAD/sd-ai/go/sdjson"
)

var (
	////go:embed testdata/compat_in1.json
	compatIn1 string
	////go:embed testdata/compat_out1.json
	compatOut1 string
)

func TestCompatTransformation(t *testing.T) {
	t.Skip("TODO: get compat example")

	var in Model
	err := json.Unmarshal([]byte(compatIn1), &in)
	require.NoError(t, err)

	actual := in.Compat()
	var expected sdjson.Model
	err = json.Unmarshal([]byte(compatOut1), &expected)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}
