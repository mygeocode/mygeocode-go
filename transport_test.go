package mygeocode

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// The standard library transport against a small local HTTP server (no internet).

func localServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body bytes.Buffer
		_, _ = body.ReadFrom(r.Body)
		if strings.HasPrefix(r.URL.Path, "/v1/slow") {
			time.Sleep(1500 * time.Millisecond)
		}
		w.Header().Set("Content-Type", "application/json")
		if strings.HasPrefix(r.URL.Path, "/v1/limited") {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
			_, _ = w.Write([]byte(`{"status":"error","error":{"code":"rate_limited","message":"Slow down."}}`))
			return
		}
		w.Header().Set("X-Request-Id", "local-1")
		w.Header().Set("X-Quota-Limit", "2500")
		out := map[string]interface{}{
			"status": "ok", "method": r.Method, "path": r.URL.RequestURI(), "ua": r.Header.Get("User-Agent"),
			"key": r.Header.Get("X-API-Key"), "content_type": r.Header.Get("Content-Type"), "body": body.String(),
			"query": nil, "results": []interface{}{},
		}
		_ = json.NewEncoder(w).Encode(out)
	}))
}

func echoed(t *testing.T, raw []byte) map[string]interface{} {
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestGetOverTheRealTransport(t *testing.T) {
	defer noEnvKey(t)()
	srv := localServer(t)
	defer srv.Close()
	c := NewClient(WithBaseURL(srv.URL+"/v1"), WithKey("mg_local"))
	r, err := c.Forward(bg, "Brandenburg Gate, Berlin", nil)
	if err != nil {
		t.Fatal(err)
	}
	m := echoed(t, r.Raw)
	if m["method"] != "GET" || m["key"] != "mg_local" || !strings.HasPrefix(m["ua"].(string), "mygeocode-go/") {
		t.Fatal(m)
	}
	if !strings.HasPrefix(m["path"].(string), "/v1/forward?") || !strings.Contains(m["path"].(string), "lang=en") {
		t.Fatal(m["path"])
	}
	if r.Meta.RequestID != "local-1" || IntValue(r.Meta.QuotaLimit) != 2500 {
		t.Fatalf("%+v", r.Meta)
	}
}

func TestPostOverTheRealTransport(t *testing.T) {
	srv := localServer(t)
	defer srv.Close()
	c := NewClient(WithBaseURL(srv.URL + "/v1"))
	r, err := c.ElevationBatch(bg, []Point{eiffel}, nil)
	if err != nil {
		t.Fatal(err)
	}
	m := echoed(t, r.Raw)
	if m["method"] != "POST" || m["content_type"] != "application/json" || m["body"] != `{"locations":[[48.8584,2.2945]]}` {
		t.Fatal(m)
	}
}

func TestTimeoutOverTheRealTransport(t *testing.T) {
	srv := localServer(t)
	defer srv.Close()
	c := NewClient(WithBaseURL(srv.URL+"/v1"), WithTimeout(200*time.Millisecond), WithMaxRetries(0))
	start := time.Now()
	err := c.send(bg, http.MethodGet, "slow", newQuery(), nil, &ElevationResponse{})
	var te *RequestTimeoutError
	if !errors.As(err, &te) {
		t.Fatalf("%T %v", err, err)
	}
	if time.Since(start) > 1400*time.Millisecond {
		t.Fatal("the timeout did not stop the request")
	}
}

func Test429OverTheRealTransport(t *testing.T) {
	srv := localServer(t)
	defer srv.Close()
	c := NewClient(WithBaseURL(srv.URL+"/v1"), WithMaxRetries(0))
	err := c.send(bg, http.MethodGet, "limited", newQuery(), nil, &ForwardResponse{})
	ae := apiErr(t, err)
	if ae.Code != "rate_limited" || ae.RetryAfter == nil || *ae.RetryAfter != time.Second {
		t.Fatalf("%+v", ae)
	}
}

func TestConnectionRefusedIsANetworkError(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	l.Close()
	s := &setup{}
	c := NewClient(WithBaseURL("http://"+addr+"/v1"), WithSleep(func(ctx context.Context, d time.Duration) error {
		s.sleeps = append(s.sleeps, d)
		return nil
	}))
	_, err = c.Reverse(bg, eiffel.Lat, eiffel.Lon, nil)
	var ne *NetworkError
	if !errors.As(err, &ne) {
		t.Fatalf("%T %v", err, err)
	}
	sleeps(t, s, time.Second, 2*time.Second)
}
