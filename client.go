// Package mygeocode is the official Go client for the My Geocode API: forward and reverse geocoding, address
// autocomplete, IP lookup, timezone, elevation and postal code lookup.
//
// The key is optional. Without one, calls use the 2,500 free requests a day of the caller's network.
//
//	c := mygeocode.NewClient()
//	res, err := c.Forward(ctx, "Brandenburg Gate, Berlin", nil)
//
// Every method takes a context first and a pointer to its own params struct last (nil for none). Results are in
// English unless you set a language with WithLang or the Lang field of a call.
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
	"strconv"
	"strings"
	"time"
)

// Version is the version of this library.
const Version = "1.1.1"

// DefaultBaseURL is the base URL of the API.
const DefaultBaseURL = "https://api.mygeocode.com/v1"

const userAgent = "mygeocode-go/" + Version + " (+https://www.mygeocode.com/sdks/)"

// Client calls the My Geocode API. It is safe for concurrent use by several goroutines.
type Client struct {
	key          string
	keyGiven     bool
	baseURL      string
	timeout      time.Duration
	maxRetries   int
	maxRetryWait time.Duration
	lang         string
	langSet      bool
	httpClient   *http.Client

	sleep func(ctx context.Context, d time.Duration) error
	now   func() time.Time
}

// Option configures a Client.
type Option func(*Client)

// WithKey sets the API key, sent in the X-API-Key header. Without this option the MYGEOCODE_KEY environment
// variable is used; with neither, no key is sent and the network's free requests apply. WithKey("") forces calls
// without a key.
func WithKey(key string) Option {
	return func(c *Client) { c.key = strings.TrimSpace(key); c.keyGiven = true }
}

// WithBaseURL replaces the base URL (default https://api.mygeocode.com/v1).
func WithBaseURL(u string) Option { return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") } }

// WithTimeout sets the time limit of each attempt, from sending the request to reading the whole answer
// (default 10 seconds). Zero or less means no limit besides the context's.
func WithTimeout(d time.Duration) Option { return func(c *Client) { c.timeout = d } }

// WithMaxRetries sets how many times a failed attempt may be retried (default 2). Zero turns retries off.
func WithMaxRetries(n int) Option {
	return func(c *Client) {
		if n < 0 {
			n = 0
		}
		c.maxRetries = n
	}
}

// WithMaxRetryWait sets the longest wait the client accepts before a retry (default 60 seconds). A 429 or 503 that
// asks for a longer wait is returned at once.
func WithMaxRetryWait(d time.Duration) Option { return func(c *Client) { c.maxRetryWait = d } }

// WithLang sets the language of the results (default "en"), sent as lang on Forward, Reverse, Autocomplete and
// their batch forms, and as language on Timezone and TimezoneBatch.
func WithLang(lang string) Option { return func(c *Client) { c.lang = lang; c.langSet = true } }

// WithTransport replaces the HTTP transport, for tests, proxies or custom TLS settings. The client still applies
// its own timeout, retries and headers.
func WithTransport(rt http.RoundTripper) Option {
	return func(c *Client) { c.httpClient = &http.Client{Transport: rt} }
}

// WithSleep replaces the function that waits between retries, so that tests do not really sleep. It must return
// early with the context's error when the context ends.
func WithSleep(f func(ctx context.Context, d time.Duration) error) Option {
	return func(c *Client) { c.sleep = f }
}

// WithClock replaces the clock used to work out the wait until X-Quota-Reset, for tests.
func WithClock(f func() time.Time) Option { return func(c *Client) { c.now = f } }

// NewClient returns a client with the given options.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:      DefaultBaseURL,
		timeout:      10 * time.Second,
		maxRetries:   2,
		maxRetryWait: 60 * time.Second,
		lang:         "en",
		httpClient:   &http.Client{},
		sleep:        sleepContext,
		now:          time.Now,
	}
	for _, o := range opts {
		o(c)
	}
	if !c.keyGiven {
		c.key = strings.TrimSpace(os.Getenv("MYGEOCODE_KEY"))
	}
	if c.httpClient == nil {
		c.httpClient = &http.Client{}
	}
	if c.sleep == nil {
		c.sleep = sleepContext
	}
	if c.now == nil {
		c.now = time.Now
	}
	return c
}

func sleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// answer is one HTTP response as it arrived.
type answer struct {
	status int
	header http.Header
	body   []byte
}

// call sends one request, retries it as documented and returns the decoded JSON object and its meta.
func (c *Client) call(ctx context.Context, method, path string, q url.Values, body interface{}) ([]byte, Meta, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var payload []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, Meta{}, &DecodeError{Err: err}
		}
		payload = b
	}
	u := c.baseURL + "/" + strings.TrimLeft(path, "/")
	if len(q) > 0 {
		u += "?" + q.Encode()
	}

	retried500 := false
	for attempt := 0; ; attempt++ {
		ans, err := c.once(ctx, method, u, payload)
		if err != nil {
			if method == http.MethodGet && attempt < c.maxRetries && ctx.Err() == nil {
				if c.sleep(ctx, backoff(attempt)) == nil {
					continue
				}
			}
			return nil, Meta{}, err
		}
		if ans.status == 429 || ans.status == 500 || ans.status == 503 {
			if wait, ok := c.retryWait(ans, attempt, retried500); ok {
				retried500 = retried500 || ans.status == 500
				if c.sleep(ctx, wait) == nil {
					continue
				}
			}
		}
		return c.parse(ans)
	}
}

// backoff is the wait before the retry that follows failed attempt number attempt (from 0): 1, 2, 4 ... seconds.
func backoff(attempt int) time.Duration {
	if attempt > 30 {
		attempt = 30
	}
	return time.Duration(1<<uint(attempt)) * time.Second
}

// serverWait is the wait the server asks for: Retry-After, or on a 429 the time until X-Quota-Reset.
func (c *Client) serverWait(ans *answer) *time.Duration {
	if n, ok := headerNumber(ans.header, "Retry-After"); ok {
		d := seconds(n)
		return &d
	}
	if ans.status == 429 {
		if n, ok := headerNumber(ans.header, "X-Quota-Reset"); ok {
			now := float64(c.now().UnixNano()) / 1e9
			d := seconds(n - now)
			return &d
		}
	}
	return nil
}

func seconds(n float64) time.Duration {
	if n <= 0 {
		return 0
	}
	if n > 1e9 {
		n = 1e9
	}
	return time.Duration(n * float64(time.Second))
}

// retryWait says whether an answer with status 429, 500 or 503 is retried, and after how long.
func (c *Client) retryWait(ans *answer, attempt int, retried500 bool) (time.Duration, bool) {
	if attempt >= c.maxRetries {
		return 0, false
	}
	switch ans.status {
	case 429:
		w := c.serverWait(ans)
		if w == nil || *w > c.maxRetryWait {
			return 0, false
		}
		return *w, true
	case 500:
		if retried500 {
			return 0, false
		}
		return time.Second, true
	case 503:
		wait := backoff(attempt)
		if w := c.serverWait(ans); w != nil {
			wait = *w
		}
		if wait > c.maxRetryWait {
			return 0, false
		}
		return wait, true
	}
	return 0, false
}

// once performs a single HTTP attempt. Any HTTP status is an answer; only a failure to get one is an error.
func (c *Client) once(ctx context.Context, method, u string, payload []byte) (*answer, error) {
	if err := ctx.Err(); err != nil {
		return nil, &NetworkError{Err: err}
	}
	actx := ctx
	if c.timeout > 0 {
		var cancel context.CancelFunc
		actx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}
	var rd io.Reader
	if payload != nil {
		rd = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(actx, method, u, rd)
	if err != nil {
		return nil, &NetworkError{Err: err}
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.key != "" {
		req.Header.Set("X-API-Key", c.key)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, transportError(ctx, actx, err)
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, transportError(ctx, actx, err)
	}
	return &answer{status: resp.StatusCode, header: resp.Header, body: buf.Bytes()}, nil
}

func transportError(ctx, actx context.Context, err error) error {
	if ctx.Err() != nil {
		// The caller cancelled, or the caller's own deadline passed.
		return &NetworkError{Err: ctx.Err()}
	}
	if errors.Is(actx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		return &RequestTimeoutError{Err: err}
	}
	var to interface{ Timeout() bool }
	if errors.As(err, &to) && to.Timeout() {
		return &RequestTimeoutError{Err: err}
	}
	return &NetworkError{Err: err}
}

// parse returns the JSON object of a successful answer, or the ApiError or DecodeError it stands for.
func (c *Client) parse(ans *answer) ([]byte, Meta, error) {
	meta := metaFrom(ans.header)
	obj, isObject := objectOf(ans.body)
	if ans.status >= 400 || (isObject && textOf(obj["status"]) == "error") {
		return nil, meta, c.apiError(ans, obj, isObject, meta)
	}
	if !isObject {
		return nil, meta, &DecodeError{
			Status:    ans.status,
			RequestID: meta.RequestID,
			Err:       errors.New("expected a JSON object, got a body that is not one"),
		}
	}
	return ans.body, meta, nil
}

func (c *Client) apiError(ans *answer, obj map[string]json.RawMessage, isObject bool, meta Meta) *ApiError {
	ae := &ApiError{
		Status:     ans.status,
		Code:       "http_error",
		Message:    "HTTP " + strconv.Itoa(ans.status),
		RequestID:  meta.RequestID,
		RetryAfter: c.serverWait(ans),
	}
	if isObject {
		ae.Body = json.RawMessage(ans.body)
		if e, ok := objectOf(obj["error"]); ok {
			if s := textOf(e["code"]); s != "" {
				ae.Code = s
			}
			if s := textOf(e["message"]); s != "" {
				ae.Message = s
			}
			ae.Param = textOf(e["param"])
		}
	}
	return ae
}

func headerNumber(h http.Header, name string) (float64, bool) {
	s := strings.TrimSpace(h.Get(name))
	if s == "" {
		return 0, false
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
		return 0, false
	}
	return n, true
}
