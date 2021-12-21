package main

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type failedContentProvider struct {
}

func (cp failedContentProvider) GetContent(userIP string, count int) ([]*ContentItem, error) {
	resp := make([]*ContentItem, count)
	return resp, errors.New("network error")
}

func TestTimeExpirationCacher_GetState(t *testing.T) {
	t.Run("not refreshed before expiration", func(t *testing.T) {
		cacher := NewTimeExpirationCacher(map[Provider]ProviderConfig{
			Provider1: {
				expiration: time.Minute * 10,
				length:     300,
				userIp:     "184.22.11.68",
				client:     failedContentProvider{},
			},
			Provider2: {
				expiration: time.Minute * 5,
				length:     100,
				userIp:     "184.22.11.68",
				client:     SampleContentProvider{Provider2},
			},
		})
		cacher.Start()
		defer cacher.Stop()
		state1 := cacher.GetState()
		state2 := cacher.GetState()
		assert.Equal(t, state1, state2)
	})
	t.Run("refreshed after expiration", func(t *testing.T) {
		cacher := NewTimeExpirationCacher(map[Provider]ProviderConfig{
			Provider1: {
				expiration: time.Millisecond * 100,
				length:     300,
				userIp:     "184.22.11.68",
				client:     SampleContentProvider{Provider1},
			},
			Provider2: {
				expiration: time.Minute * 5,
				length:     100,
				userIp:     "184.22.11.68",
				client:     SampleContentProvider{Provider2},
			},
			Provider3: {
				expiration: time.Minute * 20,
				length:     100,
				userIp:     "184.22.11.68",
				client:     SampleContentProvider{Provider3},
			},
		})
		cacher.Start()
		defer cacher.Stop()
		state1 := cacher.GetState()
		time.Sleep(time.Millisecond * 200)
		state2 := cacher.GetState()
		assert.NotEqual(t, state1, state2)
	})
	t.Run("return fails in correct cases", func(t *testing.T) {
		cacher := NewTimeExpirationCacher(map[Provider]ProviderConfig{
			Provider1: {
				expiration: time.Minute * 10,
				length:     300,
				userIp:     "184.22.11.68",
				client:     failedContentProvider{},
			},
			Provider2: {
				expiration: time.Minute * 5,
				length:     100,
				userIp:     "184.22.11.68",
				client:     SampleContentProvider{Provider2},
			},
		})
		cacher.Start()
		defer cacher.Stop()
		state := cacher.GetState()
		state.Fails(Provider1)
		assert.True(t, state.Fails(Provider1))
		assert.False(t, state.Fails(Provider2))
	})
	t.Run("return empty when index is more than have", func(t *testing.T) {
		cacher := NewTimeExpirationCacher(map[Provider]ProviderConfig{
			Provider1: {
				expiration: time.Minute * 5,
				length:     100,
				userIp:     "184.22.11.68",
				client:     SampleContentProvider{Provider1},
			},
		})
		cacher.Start()
		defer cacher.Stop()
		state := cacher.GetState()
		received := state.ContentItem(ContentAddress{
			Provider: Provider1,
			Index:    50,
		})
		assert.NotNil(t, received)
		received = state.ContentItem(ContentAddress{
			Provider: Provider1,
			Index:    500,
		})
		assert.Nil(t, received)
	})
}
