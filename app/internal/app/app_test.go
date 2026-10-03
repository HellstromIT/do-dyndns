package dyndns

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/HellstromIT/do-dyndns/app/cmd/do-dyndns/internal/config"
)

func TestGetPublicIP(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		want    string
		wantErr bool
	}{
		{name: "valid ipv4", status: http.StatusOK, body: `{"ip":"203.0.113.7"}`, want: "203.0.113.7"},
		{name: "ipv4-mapped ipv6", status: http.StatusOK, body: `{"ip":"::ffff:203.0.113.7"}`, want: "203.0.113.7"},
		{name: "ipv6", status: http.StatusOK, body: `{"ip":"2001:db8::1"}`, wantErr: true},
		{name: "hostname", status: http.StatusOK, body: `{"ip":"evil.example.com"}`, wantErr: true},
		{name: "empty ip", status: http.StatusOK, body: `{"ip":""}`, wantErr: true},
		{name: "null body", status: http.StatusOK, body: `null`, wantErr: true},
		{name: "not json", status: http.StatusOK, body: `203.0.113.7`, wantErr: true},
		{name: "error status", status: http.StatusInternalServerError, body: `{"ip":"203.0.113.7"}`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			var c config.Config
			c.Ifconfig.Host = srv.URL
			c.Ifconfig.Uri = "/"

			got, err := getPublicIP(c)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.IP != tt.want {
				t.Fatalf("got %q, want %q", got.IP, tt.want)
			}
		})
	}
}
