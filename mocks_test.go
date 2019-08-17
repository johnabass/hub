package hub

import (
	"github.com/stretchr/testify/mock"
)

type mockListener struct {
	// m is the mock.  rather than being nested, it's given its own
	// field name to ensure this enclosing type has only (1) method in
	// its method set and thus be a valid listener.
	m mock.Mock
}

func (m *mockListener) OnEvent(e TestEvent) {
	m.m.Called(e)
}
