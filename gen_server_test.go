package gen

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"
)

func TestServer(t *testing.T) {
	const num = 5
	s := newServer(context.Background())
	for i := 0; i < num; i++ {
		s.Start()
	}

	for i := 0; i < num; i++ {
		assert.Nil(t, s.Ping(10*time.Millisecond))
		err := s.Submit(func(ctx context.Context) {
			select {
			case <-ctx.Done():
			case <-time.After(20 * time.Millisecond):
				t.Error("shoud timeout")
			}
		}, Timeout(10*time.Millisecond))
		assert.Nil(t, err)
		s.Submit(func(ctx context.Context) {
			time.Sleep(2 * time.Second)
		}, Async)
		ch := make(chan bool)
		s.Submit(func(ctx context.Context) {
			ch <- true
		}, Async)
		select {
		case <-ch:
		case <-time.After(10 * time.Millisecond):
			t.Error("should async")
		}
		s.Submit(func(ctx context.Context) {
			panic("")
		})
	}

	for i := 0; i < num; i++ {
		s.Stop()
	}
}
