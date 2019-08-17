package hub

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestEvent struct {
	Value   int
	Message string
}

type BadListener struct {
}

func (bl BadListener) Invalid(a, b int) {
}

func testHubPublishSubscribe(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		neverCalled = func(*http.Request) {
			assert.Fail("This listener should never have been called")
		}

		l1 = new(mockListener)
		l2 = new(mockListener)

		l3        = make(chan TestEvent, 1)
		awaitStop = new(sync.WaitGroup)
		remote    = new(mockListener)

		firstEvent  = TestEvent{Value: 1, Message: "first"}
		secondEvent = TestEvent{Value: 2, Message: "second"}

		h = New()
	)

	require.NotNil(h)

	awaitStop.Add(1)
	go func() {
		defer awaitStop.Done()
		for e := range l3 {
			// just use another mock to verify behavior
			remote.OnEvent(e)
		}
	}()

	cancel1, err := h.Subscribe(l1)
	require.NoError(err)
	require.NotNil(cancel1)

	cancel2, err := h.Subscribe(l2.OnEvent)
	require.NoError(err)
	require.NotNil(cancel2)

	cancel3, err := h.Subscribe(l3)
	require.NoError(err)
	require.NotNil(cancel3)

	cancel4, err := h.Subscribe(neverCalled)
	require.NoError(err)
	require.NotNil(cancel4)

	l1.m.On("OnEvent", firstEvent).Once()
	l1.m.On("OnEvent", secondEvent).Once()
	l2.m.On("OnEvent", firstEvent).Once()
	remote.m.On("OnEvent", firstEvent).Once()
	remote.m.On("OnEvent", secondEvent).Once()

	h.Publish(firstEvent)
	h.Publish("a random event")

	cancel2()

	h.Publish(secondEvent)
	h.Publish("another random event")

	cancel1()
	cancel3()

	h.Publish(firstEvent)
	h.Publish(secondEvent)

	// should be idempotent
	cancel1()
	cancel2()
	cancel3()

	h.Publish(firstEvent)
	h.Publish(secondEvent)

	close(l3)
	awaitStop.Wait()

	l1.m.AssertExpectations(t)
	l2.m.AssertExpectations(t)
	remote.m.AssertExpectations(t)
}

func testHubInvalidSubscribe(t *testing.T) {
	testData := []interface{}{
		func(io.Reader) {},                       // interfaces aren't allowed
		func(a, b int) {},                        // too many inputs
		func() {},                                // no inputs
		func(a int) error { return nil },         // outputs aren't allowed
		new(bytes.Buffer),                        // only (1) method is allowed
		(<-chan TestEvent)(make(chan TestEvent)), // channels can't be receive-only
		BadListener{},                            // method isn't valid
		nil,                                      // nils aren't allowed
	}

	for i, bad := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert = assert.New(t)
				h      = New()
			)

			cancel, err := h.Subscribe(bad)
			assert.Error(err)
			assert.Nil(cancel)
		})
	}
}

func TestHub(t *testing.T) {
	t.Run("PublishSubscribe", testHubPublishSubscribe)
	t.Run("InvalidSubscribe", testHubInvalidSubscribe)
}
