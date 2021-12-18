package main

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
	Sequence(state State, limit, offset int) []ContentAddress
}

func (s Service) ContentItems(limit, offset int) (output []*ContentItem, err error) {
	// todo validation
	output = make([]*ContentItem, 0, limit)
	state := s.cacher.GetState()
	addressSequence := s.sequencer.Sequence(state, limit, offset)
	for n, address := range addressSequence {
		output[n] = state.ContentItem(address)
	}
	return
}
