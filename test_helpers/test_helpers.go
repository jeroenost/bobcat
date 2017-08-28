package test_helpers

import (
	"fmt"
	. "github.com/ThoughtWorksStudios/bobcat/common"
	"strings"
	"testing"
	"time"
)

func Assert(t *testing.T, actual bool, message string, tokens ...interface{}) {
	if !actual {
		t.Errorf(message, tokens...)
	}
}

func AssertNotNil(t *testing.T, actual interface{}, message string, tokens ...interface{}) {
	if actual == nil {
		t.Errorf(message, tokens...)
	}
}

func AssertNil(t *testing.T, actual interface{}, message string, tokens ...interface{}) {
	if actual != nil {
		t.Errorf(message, tokens...)
	}
}

func AssertTimeEqual(t *testing.T, expected, actual time.Time, optionalMessageAndTokens ...interface{}) {
	if !expected.Equal(actual) {
		failMessage := withUserMessage("Expected %v to be %v", optionalMessageAndTokens...)
		t.Errorf(failMessage, expected, actual)
	}
}

func AssertEqual(t *testing.T, expected, actual interface{}, optionalMessageAndTokens ...interface{}) {
	if expected != actual {
		failMessage := withUserMessage("Expected %v == %v", optionalMessageAndTokens...)
		t.Errorf(failMessage, expected, actual)
	}
}

func AssertNotEqual(t *testing.T, expected, actual interface{}, optionalMessageAndTokens ...interface{}) {
	if expected != actual {
		failMessage := withUserMessage("Expected %v != %v", optionalMessageAndTokens...)
		t.Errorf(failMessage, expected, actual)
	}
}

func withUserMessage(defaultMessage string, stringAndMaybeTokens ...interface{}) string {
	if len(stringAndMaybeTokens) == 0 {
		return defaultMessage
	}

	if additionalMessage, isStr := stringAndMaybeTokens[0].(string); isStr {
		if additionalMessage != "" {
			tokens := stringAndMaybeTokens[1:]
			defaultMessage = fmt.Sprintf("%s;\n\t\t%s", fmt.Sprintf(additionalMessage, tokens...), defaultMessage)
		}
	}

	return defaultMessage
}

func contains(arr []string, candidate string) bool {
	for _, v := range arr {
		if v == candidate {
			return true
		}
	}

	return false
}

func ExpectsError(t *testing.T, expectedMessage string, err error) {
	if err == nil {
		t.Errorf("Expected error [%s], but received none", expectedMessage)
		return
	}

	if !strings.Contains(err.Error(), expectedMessage) {
		t.Errorf("Failed to receive correct error message\n  expected: [%s]\n    actual: [%v]", expectedMessage, err)
	}
}

func AssertContains(t *testing.T, arr []string, candidate string) {
	if !contains(arr, candidate) {
		t.Errorf("Expected %v to contain %v, but didn't.", arr, candidate)
	}
}

type TestEmitter struct {
	result []EntityResult
}

func (te *TestEmitter) Receiver() EntityResult {
	return te.result[0]
}

func (te *TestEmitter) Emit(entity EntityResult, entityType string) error {
	te.result = append(te.result, entity)
	return nil
}

func (te *TestEmitter) NextEmitter(current EntityResult, key string, isMultiValue bool) Emitter {
	return te
}

func (te *TestEmitter) Finalize() error {
	return nil
}

func (te *TestEmitter) Shift() EntityResult {
	entity := te.result[0]
	te.result = te.result[1:]
	return entity
}

func NewTestEmitter() *TestEmitter {
	return &TestEmitter{}
}
