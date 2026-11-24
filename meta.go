package mygeocode

import (
	"math"
	"net/http"
	"strings"
)

// Meta holds the figures from the response headers. A field is nil when its header was not sent.
type Meta struct {
	// RequestID is the X-Request-Id header; quote it to support.
	RequestID string
	// QuotaLimit is X-Quota-Limit: the free requests a day.
	QuotaLimit *int64
	// QuotaUsed is X-Quota-Used: the free requests used today.
	QuotaUsed *int64
	// QuotaFreeRemaining is X-Quota-Free-Remaining: the free requests left today.
	QuotaFreeRemaining *int64
	// QuotaReset is X-Quota-Reset: the Unix time the free requests start again.
	QuotaReset *int64
	// CreditsRemaining is X-Credits-Remaining: the paid credits left, on calls with a key.
	CreditsRemaining *int64
}

func metaFrom(h http.Header) Meta {
	whole := func(name string) *int64 {
		n, ok := headerNumber(h, name)
		if !ok || math.Abs(n) > 9e18 {
			return nil
		}
		v := int64(n)
		return &v
	}
	return Meta{
		RequestID:          strings.TrimSpace(h.Get("X-Request-Id")),
		QuotaLimit:         whole("X-Quota-Limit"),
		QuotaUsed:          whole("X-Quota-Used"),
		QuotaFreeRemaining: whole("X-Quota-Free-Remaining"),
		QuotaReset:         whole("X-Quota-Reset"),
		CreditsRemaining:   whole("X-Credits-Remaining"),
	}
}
