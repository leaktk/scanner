package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAndSetLoggerLevel(t *testing.T) {
	// Default should be INFO
	assert.Equal(t, GetLoggerLevel().String(), INFO.String())

	// It should be changeable
	assert.Nil(t, SetLoggerLevel(DEBUG.String()))
	assert.Equal(t, GetLoggerLevel().String(), DEBUG.String())
	assert.Nil(t, SetLoggerLevel(INFO.String()))
	assert.Equal(t, GetLoggerLevel().String(), INFO.String())
}
