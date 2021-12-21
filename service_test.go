package main

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testSequencer struct {
	addresses []ContentAddress
	err       error
}

func (t testSequencer) Sequence(state State, limit, offset int) ([]ContentAddress, error) {
	return t.addresses, t.err
}

type testCacher struct {
	state State
}

func (t testCacher) GetState() State {
	return t.state
}

func TestService_ContentItems(t *testing.T) {
	t.Run("normal service run", func(t *testing.T) {
		s := testSequencer{
			addresses: []ContentAddress{
				{Provider: "p1", Index: 0},
				{Provider: "p2", Index: 0},
				{Provider: "p1", Index: 1},
			},
			err: nil,
		}
		c := testCacher{state: &inMemoryState{
			content: map[Provider][]*ContentItem{
				"p1": {
					&ContentItem{ID: "p1-0"},
					&ContentItem{ID: "p1-1"},
				},
				"p2": {
					&ContentItem{ID: "p2-0"},
				},
			},
		}}
		items, err := MakeService(c, s).ContentItems(10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(items))
		expected, err := json.Marshal([]*ContentItem{{ID: "p1-0"}, {ID: "p2-0"}, {ID: "p1-1"}})
		assert.NoError(t, err)
		actual, err := json.Marshal(items)
		assert.NoError(t, err)
		assert.Equal(t, string(expected), string(actual))
	})
	t.Run("there is no some content item", func(t *testing.T) {
		s := testSequencer{
			addresses: []ContentAddress{
				{Provider: "p1", Index: 0},
				{Provider: "p2", Index: 0},
				{Provider: "p1", Index: 1},
			},
			err: nil,
		}
		c := testCacher{state: &inMemoryState{
			content: map[Provider][]*ContentItem{
				"p1": {
					&ContentItem{ID: "p1-0"},
				},
				"p2": {
					&ContentItem{ID: "p2-0"},
				},
			},
		}}
		items, err := MakeService(c, s).ContentItems(10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(items))
		expected, err := json.Marshal([]*ContentItem{{ID: "p1-0"}, {ID: "p2-0"}})
		assert.NoError(t, err)
		actual, err := json.Marshal(items)
		assert.NoError(t, err)
		assert.Equal(t, string(expected), string(actual))
	})
	t.Run("sequencer returns error", func(t *testing.T) {
		s := testSequencer{
			addresses: nil,
			err:       errors.New("some error"),
		}
		c := testCacher{state: &inMemoryState{
			content: map[Provider][]*ContentItem{
				"p1": {
					&ContentItem{ID: "p1-0"},
				},
				"p2": {
					&ContentItem{ID: "p2-0"},
				},
			},
		}}
		_, err := MakeService(c, s).ContentItems(10, 0)
		assert.Error(t, err)
	})
	t.Run("validation error for limit", func(t *testing.T) {
		s := testSequencer{
			addresses: []ContentAddress{{Provider: "p1", Index: 0}},
			err:       nil,
		}
		c := testCacher{state: &inMemoryState{
			content: map[Provider][]*ContentItem{"p1": {&ContentItem{ID: "p1-0"}}},
		}}
		_, err := MakeService(c, s).ContentItems(-1, 0)
		assert.Error(t, err)
	})
	t.Run("validation error for offset", func(t *testing.T) {
		s := testSequencer{
			addresses: []ContentAddress{{Provider: "p1", Index: 0}},
			err:       nil,
		}
		c := testCacher{state: &inMemoryState{
			content: map[Provider][]*ContentItem{"p1": {&ContentItem{ID: "p1-0"}}},
		}}
		_, err := MakeService(c, s).ContentItems(10, -1)
		assert.Error(t, err)
	})
}
