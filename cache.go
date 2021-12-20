package main

import (
	"sync"
	"time"
)

// TimeExpirationCacher the component to cache the data from the providers locally and refresh it on the time basis.
type TimeExpirationCacher struct {
	providerConfigs map[Provider]ProviderConfig
	lastUpdate      map[Provider]time.Time
	state           inMemoryState
	stateLock       sync.RWMutex
	stopc           chan struct{}
	finishWG        sync.WaitGroup
}

// NewTimeExpirationCacher the constructor of the TimeExpirationCacher
func NewTimeExpirationCacher(providerConfigs map[Provider]ProviderConfig) *TimeExpirationCacher {
	// todo validation here
	cacher := &TimeExpirationCacher{
		providerConfigs: providerConfigs,
		state: inMemoryState{
			content: make(map[Provider][]*ContentItem, len(providerConfigs)),
			fails:   make(map[Provider]bool, len(providerConfigs)),
		},
		lastUpdate: make(map[Provider]time.Time, len(providerConfigs)),
		stopc:      make(chan struct{}),
	}
	return cacher
}

// ProviderConfig the configuration for the provider for the TimeExpirationCacher
type ProviderConfig struct {
	expiration time.Duration
	length     int
	userIp     string
	client     Client
}

type inMemoryState struct {
	content map[Provider][]*ContentItem
	fails   map[Provider]bool
}

// Fails returns if a given provider fails to be load.
func (ims inMemoryState) Fails(p Provider) bool {
	return ims.fails[p]
}

// ContentItem returns the content item for a given provider and index.
func (ims inMemoryState) ContentItem(addr ContentAddress) *ContentItem {
	content := ims.content[addr.Provider]
	if addr.Index >= len(content) {
		return &ContentItem{}
	}
	return content[addr.Index]
}

func (ims inMemoryState) copy() inMemoryState {
	c := inMemoryState{
		content: make(map[Provider][]*ContentItem, len(ims.content)),
		fails:   make(map[Provider]bool, len(ims.fails)),
	}
	for k, v := range ims.fails {
		c.fails[k] = v
	}
	for k, v := range ims.content {
		c.content[k] = copyContentItems(v)
	}
	return c
}

func copyContentItems(contentItems []*ContentItem) []*ContentItem {
	out := make([]*ContentItem, len(contentItems))
	for i, v := range contentItems {
		out[i] = v
	}
	return out
}

// GetState returns the state with the content items saved locally and the information if a provider fails.
func (tec *TimeExpirationCacher) GetState() State {
	tec.stateLock.RLock()
	state := tec.state.copy()
	tec.stateLock.RUnlock()
	return state
}

// Start starts the component, i.e. the routines to refresh the cache.
func (tec *TimeExpirationCacher) Start() {
	firstTimeWG := &sync.WaitGroup{}
	firstTimeWG.Add(len(tec.providerConfigs))
	tec.finishWG.Add(len(tec.providerConfigs))
	for provider, providerConfig := range tec.providerConfigs {
		p := provider
		pc := providerConfig
		go func() {
			tec.updateProvider(p, pc)
			firstTimeWG.Done()
			timer := time.NewTimer(pc.expiration)
			for {
				select {
				case <-timer.C:
					tec.updateProvider(p, pc)
				case <-tec.stopc:
					tec.finishWG.Done()
					return
				}
			}
		}()
	}
	firstTimeWG.Wait()
}

// Stop stops the component, i.e. the routines to refresh the cache.
func (tec *TimeExpirationCacher) Stop() {
	close(tec.stopc)
	tec.finishWG.Wait()
}

func (tec *TimeExpirationCacher) updateProvider(provider Provider, providerConfig ProviderConfig) {
	client := providerConfig.client
	if client != nil {
		content, err := client.GetContent(providerConfig.userIp, providerConfig.length)
		tec.stateLock.Lock()
		if err != nil {
			tec.state.fails[provider] = true
		} else {
			tec.state.fails[provider] = false
			tec.state.content[provider] = content
		}
		tec.lastUpdate[provider] = time.Now()
		tec.stateLock.Unlock()
	}
}
