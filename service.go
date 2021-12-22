package main

import (
	"fmt"
	"log"
)

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
	FailsState
	ContentItem(addr ContentAddress) *ContentItem
}

// FailsState keeps the information about the provider health.
type FailsState interface {
	Fails(p Provider) bool
}

// ContentAddress contains the information about the provider and the index of the data.
type ContentAddress struct {
	Provider Provider
	Index    int
}

// Sequencer makes the sequence of provider+index for the given input page.
type Sequencer interface {
	Sequence(state FailsState, limit, offset int) ([]ContentAddress, error)
}

// ContentItems returns the desired content items.
func (s Service) ContentItems(limit, offset int) (output []*ContentItem, err error) {
	log.Print(fmt.Sprintf("called ContentItems with parameters limit=%d, offset=%d", limit, offset))
	defer log.Print(fmt.Sprintf("finished ContentItems with parameters limit=%d, offset=%d, error: %s",
		limit, offset, err))

	if limit < 0 || offset < 0 {
		err = ValidationError("limit and offset should be positive")
		return
	}
	output = make([]*ContentItem, 0, limit)
	state := s.cacher.GetState()
	addressSequence, err := s.sequencer.Sequence(state, limit, offset)
	if err != nil {
		return nil, err
	}
	for _, address := range addressSequence {
		ci := state.ContentItem(address)
		if ci != nil {
			output = append(output, state.ContentItem(address))
		}
	}
	return
}
