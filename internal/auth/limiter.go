package auth

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"hnu-tickets/internal/platform/httpx"
)

type attemptWindow struct {
	count int
	ends  time.Time
}

// AttemptLimiter bounds authentication work per socket peer and process.
// Forwarded headers are deliberately ignored without a trusted-proxy policy.
type AttemptLimiter struct {
	mu      sync.Mutex
	windows map[string]attemptWindow
	limit   int
	window  time.Duration
	now     func() time.Time
}

func NewAttemptLimiter(limit int, window time.Duration) *AttemptLimiter {
	return &AttemptLimiter{windows: make(map[string]attemptWindow), limit: limit, window: window, now: time.Now}
}

func (l *AttemptLimiter) allow(peer string) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	key, _, err := net.SplitHostPort(peer)
	if err != nil {
		key = peer
	}
	bucket, exists := l.windows[key]
	if !exists || !now.Before(bucket.ends) {
		if !exists && len(l.windows) >= 4096 {
			for k, v := range l.windows {
				if !now.Before(v.ends) {
					delete(l.windows, k)
				}
			}
			if len(l.windows) >= 4096 {
				return false, int((l.window + time.Second - 1) / time.Second)
			}
		}
		bucket = attemptWindow{ends: now.Add(l.window)}
	}
	if bucket.count >= l.limit {
		return false, int((bucket.ends.Sub(now) + time.Second - 1) / time.Second)
	}
	bucket.count++
	l.windows[key] = bucket
	return true, 0
}

func (l *AttemptLimiter) Middleware(errors *httpx.ErrorMiddleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ok, retry := l.allow(r.RemoteAddr); !ok {
				w.Header().Set("Retry-After", strconv.Itoa(retry))
				errors.Write(w, r, httpx.NewError(http.StatusTooManyRequests, "AUTH_RATE_LIMITED", "too many authentication attempts"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
