package mygeocode

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
)

var (
	eiffel  = Point{48.8584, 2.2945}
	liberty = Point{40.6892, -74.0445}
	opera   = Point{-33.8568, 151.2153}
)

// recording is one recorded answer of the API from testdata.
type recording struct {
	Request struct {
		Method string                 `json:"method"`
		Path   string                 `json:"path"`
		Query  map[string]interface{} `json:"query"`
		Body   json.RawMessage        `json:"body"`
	} `json:"request"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    json.RawMessage   `json:"body"`
}

func load(t *testing.T, name string) *recording {
	t.Helper()
	data, err := readFile(filepath.Join("testdata", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	rec := &recording{}
	if err := json.Unmarshal(data, rec); err != nil {
		t.Fatal(err)
	}
	return rec
}

// bodyOf is the recorded answer of name decoded into plain Go values.
func bodyOf(t *testing.T, name string) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(load(t, name).Body, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

// reply is a canned answer, or an error the transport returns instead.
type reply struct {
	status int
	header http.Header
	body   []byte
	err    error
}

func replay(t *testing.T, name string) reply {
	t.Helper()
	rec := load(t, name)
	h := http.Header{}
	for k, v := range rec.Headers {
		h.Set(k, v)
	}
	return reply{status: rec.Status, header: h, body: rec.Body}
}

func respond(status int, body interface{}, headers map[string]string) reply {
	data, _ := json.Marshal(body)
	h := http.Header{"Content-Type": {"application/json"}}
	for k, v := range headers {
		h.Set(k, v)
	}
	return reply{status: status, header: h, body: data}
}

func text(status int, body string) reply {
	return reply{status: status, header: http.Header{}, body: []byte(body)}
}

func fail(err error) reply { return reply{err: err} }

// timeoutErr looks like the timeout errors of the net package.
type timeoutErr struct{}

func (timeoutErr) Error() string   { return "i/o timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return true }

// sent is a request as the transport received it.
type sent struct {
	method   string
	url      *url.URL
	header   http.Header
	body     []byte
	deadline time.Duration
	query    map[string]string
}

func (s *sent) json(t *testing.T) interface{} {
	t.Helper()
	if s.body == nil {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal(s.body, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

// recorder is an http.RoundTripper that keeps every request and answers with the queued replies in turn (the last
// one repeats).
type recorder struct {
	mu    sync.Mutex
	sent  []*sent
	queue []reply
}

func newRecorder(replies ...reply) *recorder { return &recorder{queue: replies} }

func (r *recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	s := &sent{method: req.Method, url: req.URL, header: req.Header, query: map[string]string{}}
	if dl, ok := req.Context().Deadline(); ok {
		s.deadline = time.Until(dl)
	}
	for k, v := range req.URL.Query() {
		s.query[k] = v[0]
	}
	if req.Body != nil {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(req.Body)
		s.body = buf.Bytes()
	}
	r.mu.Lock()
	r.sent = append(r.sent, s)
	item := r.queue[0]
	if len(r.queue) > 1 {
		r.queue = r.queue[1:]
	}
	r.mu.Unlock()
	if item.err != nil {
		return nil, item.err
	}
	return &http.Response{
		StatusCode: item.status,
		Header:     item.header,
		Body:       nopCloser{bytes.NewReader(item.body)},
		Request:    req,
	}, nil
}

func (r *recorder) last() *sent { return r.sent[len(r.sent)-1] }

// setup holds the waits a test client slept for.
type setup struct {
	sleeps []time.Duration
}

// client returns a client that talks to rec, never sleeps and sees the time as 1000 s after the epoch.
func (s *setup) client(rec *recorder, opts ...Option) *Client {
	base := []Option{
		WithTransport(rec),
		WithSleep(func(ctx context.Context, d time.Duration) error { s.sleeps = append(s.sleeps, d); return ctx.Err() }),
		WithClock(func() time.Time { return time.Unix(1000, 0) }),
	}
	return NewClient(append(base, opts...)...)
}

// noEnvKey clears MYGEOCODE_KEY for one test and returns the function that puts it back.
func noEnvKey(t *testing.T) func() {
	saved, had := os.LookupEnv("MYGEOCODE_KEY")
	os.Unsetenv("MYGEOCODE_KEY")
	return func() {
		if had {
			os.Setenv("MYGEOCODE_KEY", saved)
		} else {
			os.Unsetenv("MYGEOCODE_KEY")
		}
	}
}

func same(a, b string) bool {
	if a == b {
		return true
	}
	x, err1 := strconv.ParseFloat(a, 64)
	y, err2 := strconv.ParseFloat(b, 64)
	return err1 == nil && err2 == nil && x == y
}

// sentLike checks that the request equals the recorded one: method, path, every query parameter and the JSON body.
func sentLike(t *testing.T, name string, s *sent) {
	t.Helper()
	want := load(t, name).Request
	if s.method != want.Method || s.url.Path != want.Path {
		t.Fatalf("%s: sent %s %s, recorded %s %s", name, s.method, s.url.Path, want.Method, want.Path)
	}
	if len(s.query) != len(want.Query) {
		t.Fatalf("%s: query %v, recorded %v", name, s.query, want.Query)
	}
	for k, v := range want.Query {
		got, ok := s.query[k]
		if !ok || !same(got, toText(v)) {
			t.Fatalf("%s: %s = %q, recorded %v", name, k, got, v)
		}
	}
	var wantBody interface{}
	if len(want.Body) > 0 {
		_ = json.Unmarshal(want.Body, &wantBody)
	}
	if !reflect.DeepEqual(s.json(t), wantBody) {
		t.Fatalf("%s: body %s, recorded %s", name, s.body, want.Body)
	}
}

func toText(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func eqQuery(t *testing.T, got map[string]string, want map[string]string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("query\n got %v\nwant %v", got, want)
	}
}

func str(t *testing.T, p *string, want string) {
	t.Helper()
	if p == nil || *p != want {
		t.Fatalf("got %v, want %q", show(p), want)
	}
}

func num(t *testing.T, p *float64, want float64) {
	t.Helper()
	if p == nil || math.Abs(*p-want) > 1e-9 {
		t.Fatalf("got %v, want %v", show(p), want)
	}
}

func whole(t *testing.T, p *int64, want int64) {
	t.Helper()
	if p == nil || *p != want {
		t.Fatalf("got %v, want %v", show(p), want)
	}
}

func show(p interface{}) string {
	v := reflect.ValueOf(p)
	if v.IsNil() {
		return "nil"
	}
	b, _ := json.Marshal(v.Elem().Interface())
	return string(b)
}

func apiErr(t *testing.T, err error) *ApiError {
	t.Helper()
	var ae *ApiError
	if !errors.As(err, &ae) {
		t.Fatalf("want *ApiError, got %T %v", err, err)
	}
	return ae
}

func errorBody(code string) map[string]interface{} {
	return map[string]interface{}{"status": "error", "error": map[string]interface{}{"code": code, "message": "x"}}
}

func sleeps(t *testing.T, s *setup, want ...time.Duration) {
	t.Helper()
	if len(s.sleeps) != len(want) {
		t.Fatalf("slept %v, want %v", s.sleeps, want)
	}
	for i := range want {
		if s.sleeps[i] != want[i] {
			t.Fatalf("slept %v, want %v", s.sleeps, want)
		}
	}
}

var bg = context.Background()

// readFile and nopCloser stand in for helpers that are missing on Go 1.13 or deprecated later.
func readFile(name string) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var buf bytes.Buffer
	_, err = buf.ReadFrom(f)
	return buf.Bytes(), err
}

type nopCloser struct{ io.Reader }

func (nopCloser) Close() error { return nil }
