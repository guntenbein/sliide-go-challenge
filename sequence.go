package main

type ConfiguredSequencer struct {
	config ContentMix
}

func MakeConfiguredSequencer(config ContentMix) ConfiguredSequencer {
	return ConfiguredSequencer{
		config: config,
	}
}

func (sq ConfiguredSequencer) Sequence(state State, limit, offset int) (addresses []ContentAddress, err error) {
	if limit < 0 || offset < 0 {
		err = ValidationError("limit and offset should be positive")
		return
	}
	providersIndex := make(map[Provider]int, len(sq.config))
	addresses = make([]ContentAddress, 0, limit)
	var configIndex int
	for i := 0; i < offset+limit; i++ {
		config := sq.config[configIndex]
		provider := config.Type
		if state.Fails(config.Type) {
			if config.Fallback != nil && !state.Fails(*config.Fallback) {
				provider = *config.Fallback
			} else {
				return
			}
		}
		providersIndex[provider]++
		if i >= offset {
			addresses = append(addresses, ContentAddress{
				Provider: provider,
				Index:    providersIndex[provider],
			})

		}

		if configIndex >= len(sq.config)-1 {
			configIndex = 0
		} else {
			configIndex++
		}
	}
	return
}
