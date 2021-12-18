package main

import (
	"sync"
	"time"
)

type TimeExpirationCacher struct {
	providerConfigs map[Provider]ProviderConfig
	lastUpdate      map[Provider]time.Time
	state           inMemoryState
	stateLock       sync.RWMutex
	updateLock      sync.Mutex
}

func NewTimeExpirationCacher(providerConfigs map[Provider]ProviderConfig) *TimeExpirationCacher {
	// todo validation here
	cacher := &TimeExpirationCacher{
		providerConfigs: providerConfigs,
		state: inMemoryState{
			content: make(map[Provider][]*ContentItem, len(providerConfigs)),
			fails:   make(map[Provider]bool, len(providerConfigs)),
		},
		lastUpdate: make(map[Provider]time.Time, len(providerConfigs)),
	}
	return cacher
}

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

func (ims inMemoryState) Fails(p Provider) bool {
	return ims.fails[p]
}

func (ims inMemoryState) ContentItem(addr ContentAddress) *ContentItem {
	content := ims.content[addr.Provider]
	if addr.Index >= len(content) {
		return &ContentItem{}
	}
	return content[addr.Index]
}

func (ims inMemoryState) copy() inMemoryState {
	copy := inMemoryState{
		content: make(map[Provider][]*ContentItem, len(ims.content)),
		fails:   make(map[Provider]bool, len(ims.fails)),
	}
	for k, v := range ims.fails {
		copy.fails[k] = v
	}
	for k, v := range ims.content {
		copy.content[k] = copyContentItems(v)
	}
	return copy
}

func copyContentItems(contentItems []*ContentItem) []*ContentItem {
	out := make([]*ContentItem, 0, len(contentItems))
	for i, v := range contentItems {
		out[i] = v
	}
	return out
}

func (tec *TimeExpirationCacher) GetState() State {
	go tec.Update()
	tec.stateLock.RLock()
	state := tec.state.copy()
	tec.stateLock.RUnlock()
	return state
}

func (tec *TimeExpirationCacher) Update() {
	tec.updateLock.Lock()
	wg := &sync.WaitGroup{}
	for provider, providerConfig := range tec.providerConfigs {
		now := time.Now()
		updateTime := tec.lastUpdate[provider]
		if now.Sub(updateTime) > providerConfig.expiration {
			wg.Add(1)
			go tec.updateProvider(provider, providerConfig, wg)
		}
	}
	wg.Wait()
	tec.updateLock.Unlock()
}

func (tec *TimeExpirationCacher) updateProvider(provider Provider, providerConfig ProviderConfig, wg *sync.WaitGroup) {
	client := providerConfig.client
	if client != nil {
		content, err := client.GetContent(providerConfig.userIp, providerConfig.length)
		tec.stateLock.Lock()
		if err != nil {
			tec.state.fails[provider] = true
		} else {
			tec.state.content[provider] = content
		}
		tec.stateLock.Unlock()
		tec.lastUpdate[provider] = time.Now()
	}
	wg.Done()
}
