package main

type ValidationError string

func (ve ValidationError) Error() string {
	return string(ve)
}

type Service struct {
	cacher    Cacher
	sequencer Sequencer
}

func MakeService(cacher Cacher, sequencer Sequencer) Service {
	return Service{
		cacher:    cacher,
		sequencer: sequencer,
	}
}

type Cacher interface {
	GetState() State
}

type State interface {
	Fails(p Provider) bool
	ContentItem(addr ContentAddress) *ContentItem
}

type ContentAddress struct {
	Provider Provider
	Index    int
}

type Sequencer interface {
	Sequence(state State, limit, offset int) ([]ContentAddress, error)
}

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
