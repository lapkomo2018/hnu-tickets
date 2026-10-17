package auth

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"hnu-tickets/internal/platform/httpx"
)

func TestLimiterRejectsBurstAndIgnoresForwardedIdentity(t *testing.T) {
	now := time.Now()
	l := NewAttemptLimiter(2, time.Minute)
	l.now = func() time.Time { return now }
	e := httpx.NewErrorMiddleware(slog.New(slog.NewTextHandler(io.Discard, nil)))
	h := l.Middleware(e)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) }))
	for i, want := range []int{204, 204, 429} {
		r := httptest.NewRequestWithContext(t.Context(), "POST", "/login", nil)
		r.RemoteAddr = "192.0.2.1:1234"
		r.Header.Set("X-Forwarded-For", string(rune('a'+i)))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		require.Equal(t, want, w.Code)
		if want == 429 {
			require.Equal(t, "60", w.Header().Get("Retry-After"))
		}
	}
	now = now.Add(time.Minute)
	ok, _ := l.allow("192.0.2.1:9876")
	require.True(t, ok)
}

func TestLimiterIsAtomicAndBounded(t *testing.T) {
	l := NewAttemptLimiter(5, time.Minute)
	var wg sync.WaitGroup
	var accepted atomic.Int32
	for range 30 {
		wg.Go(func() {
			if ok, _ := l.allow("192.0.2.1:1"); ok {
				accepted.Add(1)
			}
		})
	}
	wg.Wait()
	require.Equal(t, int32(5), accepted.Load())
	for i := range 4096 {
		l.windows[string(rune(i+1000))] = attemptWindow{ends: time.Now().Add(time.Hour)}
	}
	ok, _ := l.allow("new-address")
	require.False(t, ok)
}
