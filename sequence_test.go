package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfiguredSequencer_Sequence(t *testing.T) {
	t.Run("normal sequencer functioning, offset 0", func(t *testing.T) {
		state := inMemoryState{fails: map[Provider]bool{Provider1: false, Provider2: false, Provider3: false}}
		config := DefaultConfig
		sequencer := MakeConfiguredSequencer(config)
		addresses, err := sequencer.Sequence(state, 8, 0)
		assert.NoError(t, err)
		expected := []ContentAddress{
			{Provider: Provider1, Index: 0},
			{Provider: Provider1, Index: 1},
			{Provider: Provider2, Index: 0},
			{Provider: Provider3, Index: 0},
			{Provider: Provider1, Index: 2},
			{Provider: Provider1, Index: 3},
			{Provider: Provider1, Index: 4},
			{Provider: Provider2, Index: 1},
		}
		assert.Equal(t, expected, addresses)
	})
	t.Run("normal sequencer functioning, offset non 0", func(t *testing.T) {
		state := inMemoryState{fails: map[Provider]bool{Provider1: false, Provider2: false, Provider3: false}}
		config := DefaultConfig
		sequencer := MakeConfiguredSequencer(config)
		addresses, err := sequencer.Sequence(state, 8, 8)
		assert.NoError(t, err)
		expected := []ContentAddress{
			{Provider: Provider1, Index: 5},
			{Provider: Provider1, Index: 6},
			{Provider: Provider2, Index: 2},
			{Provider: Provider3, Index: 1},
			{Provider: Provider1, Index: 7},
			{Provider: Provider1, Index: 8},
			{Provider: Provider1, Index: 9},
			{Provider: Provider2, Index: 3},
		}
		assert.Equal(t, expected, addresses)
	})
	t.Run("a provider fails", func(t *testing.T) {
		state := inMemoryState{fails: map[Provider]bool{Provider1: false, Provider2: true, Provider3: false}}
		config := DefaultConfig
		sequencer := MakeConfiguredSequencer(config)
		addresses, err := sequencer.Sequence(state, 8, 0)
		assert.NoError(t, err)
		expected := []ContentAddress{
			{Provider: Provider1, Index: 0},
			{Provider: Provider1, Index: 1},
			{Provider: Provider3, Index: 0},
			{Provider: Provider3, Index: 1},
			{Provider: Provider1, Index: 2},
			{Provider: Provider1, Index: 3},
			{Provider: Provider1, Index: 4},
			{Provider: Provider3, Index: 2},
		}
		assert.Equal(t, expected, addresses)
	})
	t.Run("a provider fails, fallback fails", func(t *testing.T) {
		state := inMemoryState{fails: map[Provider]bool{Provider1: false, Provider2: true, Provider3: true}}
		config := DefaultConfig
		sequencer := MakeConfiguredSequencer(config)
		addresses, err := sequencer.Sequence(state, 8, 0)
		assert.NoError(t, err)
		expected := []ContentAddress{
			{Provider: Provider1, Index: 0},
			{Provider: Provider1, Index: 1},
		}
		assert.Equal(t, expected, addresses)
	})
	t.Run("a provider fails, fallback nil", func(t *testing.T) {
		state := inMemoryState{fails: map[Provider]bool{Provider1: true, Provider2: false, Provider3: false}}
		config := DefaultConfig
		sequencer := MakeConfiguredSequencer(config)
		addresses, err := sequencer.Sequence(state, 8, 0)
		assert.NoError(t, err)
		expected := []ContentAddress{
			{Provider: Provider2, Index: 0},
			{Provider: Provider2, Index: 1},
			{Provider: Provider2, Index: 2},
			{Provider: Provider3, Index: 0},
		}
		assert.Equal(t, expected, addresses)
	})
	t.Run("a provider fails, fallback fails, next page", func(t *testing.T) {
		state := inMemoryState{fails: map[Provider]bool{Provider1: false, Provider2: true, Provider3: true}}
		config := DefaultConfig
		sequencer := MakeConfiguredSequencer(config)
		addresses, err := sequencer.Sequence(state, 8, 8)
		assert.NoError(t, err)
		expected := []ContentAddress{}
		assert.Equal(t, expected, addresses)
	})
	t.Run("incorrect limit error", func(t *testing.T) {
		state := inMemoryState{fails: map[Provider]bool{Provider1: false, Provider2: false, Provider3: false}}
		config := DefaultConfig
		sequencer := MakeConfiguredSequencer(config)
		_, err := sequencer.Sequence(state, -1, 8)
		assert.Error(t, err)
	})
	t.Run("incorrect offset error", func(t *testing.T) {
		state := inMemoryState{fails: map[Provider]bool{Provider1: false, Provider2: false, Provider3: false}}
		config := DefaultConfig
		sequencer := MakeConfiguredSequencer(config)
		_, err := sequencer.Sequence(state, -1, 8)
		assert.Error(t, err)
	})
}
