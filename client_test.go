package mygeocode

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestKeylessSendsNoKeyAndOurUserAgent(t *testing.T) {
	defer noEnvKey(t)()
	rec := newRecorder(replay(t, "forward"))
	s := &setup{}
	if _, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil); err != nil {
		t.Fatal(err)
	}
	if _, has := rec.last().header["X-Api-Key"]; has {
		t.Fatal("a key was sent")
	}
	if ua := rec.last().header.Get("User-Agent"); ua != "mygeocode-go/"+Version+" (+https://www.mygeocode.com/sdks/)" {
		t.Fatalf("User-Agent %q", ua)
	}
	if u := rec.last().url.String(); !strings.HasPrefix(u, "https://api.mygeocode.com/v1/forward?") {
		t.Fatalf("url %s", u)
	}
	if rec.last().header.Get("Accept") != "application/json" {
		t.Fatal("no Accept header")
	}
}

func TestKeyGoesInTheHeaderNeverTheURL(t *testing.T) {
	defer noEnvKey(t)()
	rec := newRecorder(replay(t, "forward"))
	s := &setup{}
	if _, err := s.client(rec, WithKey("mg_test_key")).Forward(bg, "Brandenburg Gate, Berlin", nil); err != nil {
		t.Fatal(err)
	}
	if rec.last().header.Get("X-API-Key") != "mg_test_key" {
		t.Fatal("key header missing")
	}
	if strings.Contains(rec.last().url.String(), "mg_test_key") {
		t.Fatal("key in the URL")
	}
}

func TestKeyFromTheEnvironmentAndEmptyKeyMeansNone(t *testing.T) {
	defer noEnvKey(t)()
	os.Setenv("MYGEOCODE_KEY", "mg_env_key")
	rec := newRecorder(replay(t, "forward"))
	s := &setup{}
	if _, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil); err != nil {
		t.Fatal(err)
	}
	if rec.last().header.Get("X-API-Key") != "mg_env_key" {
		t.Fatal("environment key not sent")
	}
	if _, err := s.client(rec, WithKey("")).Forward(bg, "Brandenburg Gate, Berlin", nil); err != nil {
		t.Fatal(err)
	}
	if _, has := rec.last().header["X-Api-Key"]; has {
		t.Fatal("WithKey(\"\") still sent a key")
	}
}

func TestDefaultsAreTenSecondsAndEnglish(t *testing.T) {
	rec := newRecorder(replay(t, "forward"))
	s := &setup{}
	if _, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil); err != nil {
		t.Fatal(err)
	}
	if d := rec.last().deadline; d < 9*time.Second || d > 10*time.Second {
		t.Fatalf("deadline %v", d)
	}
	if rec.last().query["lang"] != "en" {
		t.Fatal("lang is not en")
	}
}

func TestBaseURLAndTimeoutReachTheTransport(t *testing.T) {
	rec := newRecorder(replay(t, "forward"))
	s := &setup{}
	c := s.client(rec, WithBaseURL("https://example.test/v1/"), WithTimeout(3*time.Second))
	if _, err := c.Forward(bg, "Brandenburg Gate, Berlin", nil); err != nil {
		t.Fatal(err)
	}
	if u := rec.last().url.String(); !strings.HasPrefix(u, "https://example.test/v1/forward?") {
		t.Fatalf("url %s", u)
	}
	if d := rec.last().deadline; d < 2*time.Second || d > 3*time.Second {
		t.Fatalf("deadline %v", d)
	}
}

func TestLangOnTheClientAndPerCall(t *testing.T) {
	rec := newRecorder(replay(t, "forward"))
	s := &setup{}
	c := s.client(rec, WithLang("de"))
	if _, err := c.Forward(bg, "Brandenburg Gate, Berlin", nil); err != nil {
		t.Fatal(err)
	}
	if rec.last().query["lang"] != "de" {
		t.Fatal("client lang not sent")
	}
	if _, err := c.Forward(bg, "Brandenburg Gate, Berlin", &ForwardParams{Lang: "fr"}); err != nil {
		t.Fatal(err)
	}
	if rec.last().query["lang"] != "fr" {
		t.Fatal("call lang not sent")
	}
}

func TestTimezoneLanguageOnlyWhenChosen(t *testing.T) {
	rec := newRecorder(replay(t, "timezone"))
	s := &setup{}
	c := s.client(rec)
	check := func(p *TimezoneParams, c *Client, want string) {
		t.Helper()
		if _, err := c.Timezone(bg, 35.6762, 139.6503, p); err != nil {
			t.Fatal(err)
		}
		got, has := rec.last().query["language"]
		if (want == "") == has || got != want {
			t.Fatalf("language %q (sent %v), want %q", got, has, want)
		}
		if _, has := rec.last().query["lang"]; has {
			t.Fatal("timezone sent lang")
		}
	}
	check(nil, c, "")
	check(&TimezoneParams{Lang: "ja"}, c, "ja")
	check(&TimezoneParams{Language: "de"}, c, "de")
	check(nil, s.client(rec, WithLang("en")), "en")
}

func TestErrorEnvelopeBecomesApiError(t *testing.T) {
	rec := newRecorder(replay(t, "error_invalid_request"))
	s := &setup{}
	_, err := s.client(rec).Forward(bg, "", nil)
	ae := apiErr(t, err)
	r := load(t, "error_invalid_request")
	var body struct {
		Error struct{ Code, Message, Param string }
	}
	_ = json.Unmarshal(r.Body, &body)
	if ae.Status != r.Status || ae.Code != body.Error.Code || ae.Message != body.Error.Message || ae.Param != body.Error.Param {
		t.Fatalf("%+v", ae)
	}
	if ae.RequestID != r.Headers["x-request-id"] || ae.RetryAfter != nil {
		t.Fatalf("%+v", ae)
	}
	var got, want interface{}
	_ = json.Unmarshal(ae.Body, &got)
	_ = json.Unmarshal(r.Body, &want)
	if !reflect.DeepEqual(got, want) {
		t.Fatal("Body is not the answer")
	}
	var base MyGeocodeError
	if !errors.As(err, &base) {
		t.Fatal("not a MyGeocodeError")
	}
	if !strings.Contains(err.Error(), ae.Code) {
		t.Fatal(err.Error())
	}
}

func TestErrorWithoutAJSONBody(t *testing.T) {
	rec := newRecorder(text(502, "<html>Bad gateway</html>"))
	s := &setup{}
	_, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil)
	ae := apiErr(t, err)
	if ae.Status != 502 || ae.Code != "http_error" || ae.Message != "HTTP 502" || ae.Body != nil {
		t.Fatalf("%+v", ae)
	}
}

func TestAnOKAnswerWithStatusErrorIsAnApiError(t *testing.T) {
	rec := newRecorder(respond(200, errorBody("invalid_request"), nil))
	s := &setup{}
	_, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil)
	if ae := apiErr(t, err); ae.Status != 200 || ae.Code != "invalid_request" {
		t.Fatalf("%+v", ae)
	}
}

func TestSuccessThatIsNotJSONIsADecodeError(t *testing.T) {
	for _, body := range []string{"not json", "[1,2]", ""} {
		rec := newRecorder(text(200, body))
		s := &setup{}
		_, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil)
		var de *DecodeError
		if !errors.As(err, &de) || de.Status != 200 {
			t.Fatalf("%q: %T %v", body, err, err)
		}
	}
}

func Test429WaitsForRetryAfter(t *testing.T) {
	rec := newRecorder(respond(429, errorBody("rate_limited"), map[string]string{"Retry-After": "7"}), replay(t, "forward"))
	s := &setup{}
	r, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil)
	if err != nil {
		t.Fatal(err)
	}
	sleeps(t, s, 7*time.Second)
	str(t, r.Status, "ok")
	if len(rec.sent) != 2 {
		t.Fatal(len(rec.sent))
	}
}

func Test429WaitsUntilTheQuotaReset(t *testing.T) {
	rec := newRecorder(respond(429, errorBody("quota_exceeded"), map[string]string{"X-Quota-Reset": "1030"}), replay(t, "forward"))
	s := &setup{}
	if _, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil); err != nil {
		t.Fatal(err)
	}
	sleeps(t, s, 30*time.Second)
}

func TestQuotaGoneForTheDayReturnsAtOnce(t *testing.T) {
	rec := newRecorder(respond(429, errorBody("quota_exceeded"), map[string]string{"X-Quota-Reset": "90000"}))
	s := &setup{}
	_, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil)
	ae := apiErr(t, err)
	if ae.Code != "quota_exceeded" || ae.RetryAfter == nil || *ae.RetryAfter != 89000*time.Second {
		t.Fatalf("%+v", ae)
	}
	sleeps(t, s)
	if len(rec.sent) != 1 {
		t.Fatal(len(rec.sent))
	}
}

func Test429WithoutAnyWaitReturnsAtOnce(t *testing.T) {
	rec := newRecorder(respond(429, errorBody("rate_limited"), nil))
	s := &setup{}
	_, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil)
	if ae := apiErr(t, err); ae.Status != 429 || ae.RetryAfter != nil {
		t.Fatalf("%+v", ae)
	}
	if len(rec.sent) != 1 {
		t.Fatal(len(rec.sent))
	}
}

func TestMaxRetryWaitIsConfigurable(t *testing.T) {
	rec := newRecorder(respond(429, errorBody("rate_limited"), map[string]string{"Retry-After": "90"}), replay(t, "forward"))
	s := &setup{}
	if _, err := s.client(rec, WithMaxRetryWait(120*time.Second)).Forward(bg, "Brandenburg Gate, Berlin", nil); err != nil {
		t.Fatal(err)
	}
	sleeps(t, s, 90*time.Second)
}

func TestRetriesStopAtMaxRetries(t *testing.T) {
	rec := newRecorder(respond(503, errorBody("unavailable"), nil))
	s := &setup{}
	_, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil)
	apiErr(t, err)
	sleeps(t, s, time.Second, 2*time.Second)
	if len(rec.sent) != 3 {
		t.Fatal(len(rec.sent))
	}
}

func TestMaxRetriesZeroNeverRetries(t *testing.T) {
	rec := newRecorder(respond(503, errorBody("unavailable"), nil))
	s := &setup{}
	_, err := s.client(rec, WithMaxRetries(0)).Forward(bg, "Brandenburg Gate, Berlin", nil)
	apiErr(t, err)
	if len(rec.sent) != 1 {
		t.Fatal(len(rec.sent))
	}
}

func Test503UsesRetryAfter(t *testing.T) {
	rec := newRecorder(respond(503, errorBody("maintenance"), map[string]string{"Retry-After": "5"}), replay(t, "forward"))
	s := &setup{}
	if _, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil); err != nil {
		t.Fatal(err)
	}
	sleeps(t, s, 5*time.Second)
}

func Test503WithALongRetryAfterReturns(t *testing.T) {
	rec := newRecorder(respond(503, errorBody("maintenance"), map[string]string{"Retry-After": "900"}))
	s := &setup{}
	_, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil)
	if ae := apiErr(t, err); ae.RetryAfter == nil || *ae.RetryAfter != 900*time.Second {
		t.Fatalf("%+v", ae)
	}
	sleeps(t, s)
}

func Test500IsRetriedOnceAfterOneSecond(t *testing.T) {
	broken := respond(500, errorBody("server_error"), nil)
	rec := newRecorder(broken, replay(t, "forward"))
	s := &setup{}
	if _, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil); err != nil {
		t.Fatal(err)
	}
	sleeps(t, s, time.Second)
	rec = newRecorder(broken)
	s = &setup{}
	_, err := s.client(rec, WithMaxRetries(5)).Forward(bg, "Brandenburg Gate, Berlin", nil)
	apiErr(t, err)
	if len(rec.sent) != 2 {
		t.Fatal(len(rec.sent))
	}
	sleeps(t, s, time.Second)
}

func Test500OnPostAndMaxRetries(t *testing.T) {
	broken := respond(500, errorBody("server_error"), nil)
	rec := newRecorder(broken, replay(t, "forward_batch"))
	s := &setup{}
	if _, err := s.client(rec).ForwardBatch(bg, []string{"Brandenburg Gate, Berlin"}, nil); err != nil {
		t.Fatal(err)
	}
	if len(rec.sent) != 2 {
		t.Fatal(len(rec.sent))
	}
	rec = newRecorder(broken)
	_, err := s.client(rec, WithMaxRetries(0)).ForwardBatch(bg, []string{"Brandenburg Gate, Berlin"}, nil)
	apiErr(t, err)
	if len(rec.sent) != 1 {
		t.Fatal(len(rec.sent))
	}
}

func TestConnectionErrorsAreRetriedOnGet(t *testing.T) {
	refused := fail(errors.New("connection refused"))
	rec := newRecorder(refused, refused, replay(t, "forward"))
	s := &setup{}
	if _, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil); err != nil {
		t.Fatal(err)
	}
	sleeps(t, s, time.Second, 2*time.Second)
}

func TestNetworkErrorAfterTheRetries(t *testing.T) {
	rec := newRecorder(fail(errors.New("connection refused")))
	s := &setup{}
	_, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil)
	var ne *NetworkError
	if !errors.As(err, &ne) {
		t.Fatalf("%T %v", err, err)
	}
	if len(rec.sent) != 3 {
		t.Fatal(len(rec.sent))
	}
}

func TestTimeoutsAreRetriedThenReturned(t *testing.T) {
	rec := newRecorder(fail(timeoutErr{}))
	s := &setup{}
	_, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil)
	var te *RequestTimeoutError
	if !errors.As(err, &te) || !te.Timeout() {
		t.Fatalf("%T %v", err, err)
	}
	if len(rec.sent) != 3 {
		t.Fatal(len(rec.sent))
	}
}

func TestPostIsNotRetriedAfterAConnectionError(t *testing.T) {
	rec := newRecorder(fail(errors.New("connection reset")), replay(t, "forward_batch"))
	s := &setup{}
	_, err := s.client(rec).ForwardBatch(bg, []string{"Brandenburg Gate, Berlin"}, nil)
	var ne *NetworkError
	if !errors.As(err, &ne) {
		t.Fatalf("%T %v", err, err)
	}
	if len(rec.sent) != 1 {
		t.Fatal(len(rec.sent))
	}
}

func TestPostIsRetriedOn429(t *testing.T) {
	rec := newRecorder(respond(429, errorBody("rate_limited"), map[string]string{"Retry-After": "2"}), replay(t, "forward_batch"))
	s := &setup{}
	if _, err := s.client(rec).ForwardBatch(bg, []string{"Brandenburg Gate, Berlin"}, nil); err != nil {
		t.Fatal(err)
	}
	sleeps(t, s, 2*time.Second)
}

func TestACancelledContextStopsTheRetries(t *testing.T) {
	ctx, cancel := context.WithCancel(bg)
	rec := newRecorder(respond(503, errorBody("unavailable"), nil))
	c := NewClient(WithTransport(rec), WithSleep(func(ctx context.Context, d time.Duration) error {
		cancel()
		return ctx.Err()
	}))
	_, err := c.Forward(ctx, "Brandenburg Gate, Berlin", nil)
	apiErr(t, err)
	if len(rec.sent) != 1 {
		t.Fatal(len(rec.sent))
	}
	_, err = c.Forward(ctx, "Brandenburg Gate, Berlin", nil)
	var ne *NetworkError
	if !errors.As(err, &ne) || !errors.Is(err, context.Canceled) {
		t.Fatalf("%T %v", err, err)
	}
}
