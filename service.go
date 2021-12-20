package main

type ValidationError string

func (ve ValidationError) Error() string {
	return string(ve)
}

// Service the service to provide the data for the given config.
type Service struct {
	cacher    Cacher
	sequencer Sequencer
}

// MakeService is a constructor for the Service, it has the checher component and the sequencer component as the input.
func MakeService(cacher Cacher, sequencer Sequencer) Service {
	return Service{
		cacher:    cacher,
		sequencer: sequencer,
	}
}

// Cacher is responsible for keeping the state and providing it to the service on request.
type Cacher interface {
	GetState() State
}

// State keeps the desired content iteems and information about the provider health.
type State interface {
	Fails(p Provider) bool
	ContentItem(addr ContentAddress) *ContentItem
}

// ContentAddress contains the information about the provider and the index of the data.
type ContentAddress struct {
	Provider Provider
	Index    int
}

// Sequencer makes the sequence of provider+index for the given input page.
type Sequencer interface {
	Sequence(state State, limit, offset int) ([]ContentAddress, error)
}

// ContentItems returns the desired content items.
func (s Service) ContentItems(limit, offset int) (output []*ContentItem, err error) {
	if limit < 0 || offset < 0 {
		err = ValidationError("limit and offset should be positive")
		return
	}
	output = make([]*ContentItem, limit)
	state := s.cacher.GetState()
	addressSequence, err := s.sequencer.Sequence(state, limit, offset)
	if err != nil {
		return nil, err
	}
	for n, address := range addressSequence {
		output[n] = state.ContentItem(address)
	}
	return
}
