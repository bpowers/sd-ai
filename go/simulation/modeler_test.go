package simulation

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/response.json
var responseIn1 string

func TestResponseUnmarshal(t *testing.T) {
	var in Model
	err := json.Unmarshal([]byte(responseIn1), &in)
	require.NoError(t, err)
}
