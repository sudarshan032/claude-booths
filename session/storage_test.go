package session

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type memState struct{ data json.RawMessage }

func (m *memState) SaveInstances(d json.RawMessage) error { m.data = d; return nil }
func (m *memState) GetInstances() json.RawMessage         { return m.data }
func (m *memState) DeleteAllInstances() error             { m.data = json.RawMessage("[]"); return nil }

// Deleting used to rebuild every stored booth (reattaching to each tmux session); a booth
// that failed to restore made the delete fail and the deleted booth stayed.
func TestDeleteInstanceWorksOnStoredData(t *testing.T) {
	st := &memState{data: json.RawMessage(`[{"title":"a","status":1,"program":"x"},{"title":"b","status":1,"program":"x"}]`)}
	s, err := NewStorage(st)
	require.NoError(t, err)

	require.NoError(t, s.DeleteInstance("a"))
	var left []InstanceData
	require.NoError(t, json.Unmarshal(st.data, &left))
	require.Len(t, left, 1)
	require.Equal(t, "b", left[0].Title)

	require.True(t, errors.Is(s.DeleteInstance("nope"), ErrInstanceNotFound))
}
