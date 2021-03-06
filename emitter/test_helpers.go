package emitter

import . "github.com/ThoughtWorksStudios/bobcat/common"

/**
 * holds flat, in-memory list of EntityResults for testing and
 * has convenience method to inspect results
 */
type TestEmitter struct {
	result []EntityResult
	closed bool
}

func (te *TestEmitter) Receiver() EntityStore {
	return EntityResult{}
}

func (te *TestEmitter) Emit(entity EntityResult, entityType string) error {
	te.result = append(te.result, entity)
	return nil
}

func (te *TestEmitter) NextEmitter(current EntityStore, key string, isMultiValue bool) Emitter {
	return te
}

func (te *TestEmitter) Init() error {
	return nil
}

func (te *TestEmitter) Finalize() error {
	te.closed = true
	return nil
}

func (te *TestEmitter) Closed() bool {
	return te.closed
}

func (te *TestEmitter) Shift() EntityResult {
	entity := te.result[0]
	te.result = te.result[1:]
	return entity
}

// literally does nothing; here just to satisfy params
type DummyEmitter struct{}

func (de *DummyEmitter) Receiver() EntityStore {
	return EntityResult{}
}

func (de *DummyEmitter) Emit(entity EntityResult, entityType string) error {
	return nil
}

func (de *DummyEmitter) NextEmitter(current EntityStore, key string, isMultiValue bool) Emitter {
	return de
}

func (de *DummyEmitter) Init() error {
	return nil
}

func (de *DummyEmitter) Finalize() error {
	return nil
}

// test helpers

func NewTestEmitter() *TestEmitter {
	return &TestEmitter{result: make([]EntityResult, 0)}
}

func NewDummyEmitter() *DummyEmitter {
	return &DummyEmitter{}
}

type TestProvider struct{}

func (p *TestProvider) Get(key string) (Emitter, error) {
	return NewTestEmitter(), nil
}
