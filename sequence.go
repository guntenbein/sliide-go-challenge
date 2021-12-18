package main

type ConfiguredSequencer struct {
	config ContentMix
}

func MakeConfiguredSequencer(config ContentMix) ConfiguredSequencer {
	return ConfiguredSequencer{
		config: config,
	}
}

func (sq ConfiguredSequencer) Sequence(state State, limit, offset int) (addresses []ContentAddress) {
	// todo validation here
	providersIndex := make(map[Provider]int, len(sq.config))
	addresses = make([]ContentAddress, 0, limit)
	for i := 0; i < offset+limit; {
		for _, config := range sq.config {
			provider := config.Type
			if state.Fails(config.Type) {
				if config.Fallback != nil && !state.Fails(*config.Fallback) {
					provider = *config.Fallback
				} else {
					return
				}
			}
			if i >= limit {
				addresses = append(addresses, ContentAddress{
					Provider: provider,
					Index:    providersIndex[provider],
				})
			}
			providersIndex[config.Type]++
			i++
		}
	}
	return
}
