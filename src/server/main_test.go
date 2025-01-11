package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefinedTogether(t *testing.T) {
	assert.True(t, definedTogether())
	assert.True(t, definedTogether("", "", ""))
	assert.True(t, definedTogether("big", "small", "micro"))

	assert.False(t, definedTogether("", "blah", "okay"))
	assert.False(t, definedTogether("nope", "blah", ""))
}
