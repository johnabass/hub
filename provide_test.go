package hub

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvide(t *testing.T) {
	assert := assert.New(t)
	p, s, h := Provide()
	assert.NotNil(p)
	assert.NotNil(s)
	assert.NotNil(h)
}
