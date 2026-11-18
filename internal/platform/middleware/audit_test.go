package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestAuditIdentityAndSecretExclusion(t *testing.T) {
	var out bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&out, nil)))
	defer slog.SetDefault(old)
	router := chi.NewRouter()
	router.Use(RequestID, RequestLogger)
	router.Post("/tickets/{id}", func(w http.ResponseWriter, r *http.Request) {
		SetAuditActor(r, "verified-user", "USER")
		w.WriteHeader(http.StatusForbidden)
	})
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/tickets/123e4567-e89b-12d3-a456-426614174000?password=SECRET", strings.NewReader("SECRET"))
	request.Header.Set("Authorization", "Bearer SECRET")
	request.Header.Set("X-Request-ID", "SECRET")
	router.ServeHTTP(httptest.NewRecorder(), request)
	if strings.Contains(out.String(), "SECRET") {
		t.Fatal("secret leaked")
	}
	var event map[string]any
	if err := json.Unmarshal(out.Bytes(), &event); err != nil {
		t.Fatal(err)
	}
	if event["actor_id"] != "verified-user" || event["status"] != float64(403) || event["route"] != "/tickets/{id}" || event["level"] != "WARN" {
		t.Fatalf("unexpected event: %v", event)
	}
	if event["object_id"] != "123e4567-e89b-12d3-a456-426614174000" {
		t.Fatal("object missing")
	}
}
