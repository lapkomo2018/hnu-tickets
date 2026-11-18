package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

type auditKey struct{}
type auditActor struct{ ID, Role string }

// SetAuditActor records an identity only after authentication.
func SetAuditActor(r *http.Request, id, role string) {
	if actor, ok := r.Context().Value(auditKey{}).(*auditActor); ok {
		actor.ID, actor.Role = id, role
	}
}
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		actor := &auditActor{ID: "anonymous"}
		r = r.WithContext(context.WithValue(r.Context(), auditKey{}, actor))
		recorder := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(recorder, r)
		status := recorder.Status()
		if status == 0 {
			status = http.StatusOK
		}
		route, object := "unmatched", ""
		if ctx := chi.RouteContext(r.Context()); ctx != nil {
			if ctx.RoutePattern() != "" {
				route = ctx.RoutePattern()
			}
			for _, value := range ctx.URLParams.Values {
				if id, err := uuid.Parse(value); err == nil {
					object = id.String()
					break
				}
			}
		}
		peer, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			peer = "unknown"
		}
		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}
		slog.Log(r.Context(), level, "http.operation", "request_id", recorder.Header().Get("X-Request-ID"),
			"actor_id", actor.ID, "role", actor.Role, "method", r.Method, "route", route, "object_id", object,
			"peer", peer, "status", status, "bytes", recorder.BytesWritten(), "duration_ms", time.Since(started).Milliseconds())
	})
}

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if _, ok := allowed[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
