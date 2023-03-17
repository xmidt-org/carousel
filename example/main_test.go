package example

import (
	"plugin"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xmidt-org/carousel/pkg/carousel"
)

func TestCheckHost(t *testing.T) {
	assert := assert.New(t)
	assert.True(CheckHost("example2.testing.com"))
	assert.False(CheckHost("example.testing.com"))
	assert.False(CheckHost("127.0.0.1"))
	assert.False(CheckHost("localhost"))
	assert.True(CheckHost("example32.com"))
}

func TestCheckHostType(t *testing.T) {
	assert := assert.New(t)
	var sys plugin.Symbol
	sys = CheckHost
	if _, ok := sys.(func(fqdn string) bool); !ok {
		assert.Fail("Check host is not a func(fqdn string) bool")
	}
	sys = Check
	if _, ok := sys.(carousel.HostValidator); !ok {
		assert.Fail("Check is not a carousel.HostValidator")
	}
}
