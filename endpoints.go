package mygeocode

import (
	"context"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// Point is a latitude and longitude in decimal degrees.
type Point struct {
	Lat float64
	Lon float64
}

// BBox is a box in decimal degrees, sent as south,west,north,east.
type BBox struct {
	South float64
	West  float64
	North float64
	East  float64
}

// Params holds extra query parameters, sent as they are, so that parameters added to the API later work without a
// new release. A name in Params replaces the value the library would have sent.
type Params map[string]string

// ---- encoding ---------------------------------------------------------------------------------------------------

type query url.Values

func (q query) text(name, v string) {
	if v != "" {
		url.Values(q).Set(name, v)
	}
}

func (q query) whole(name string, v int64) {
	if v != 0 {
		url.Values(q).Set(name, strconv.FormatInt(v, 10))
	}
}

func (q query) flag(name string, v bool) {
	if v {
		url.Values(q).Set(name, "1")
	}
}

func (q query) list(name string, v []string) { q.text(name, strings.Join(v, ",")) }

func (q query) point(name string, p *Point) {
	if p != nil {
		q.text(name, pair(*p))
	}
}

func (q query) box(name string, b *BBox) {
	if b != nil {
		q.text(name, number(b.South)+","+number(b.West)+","+number(b.North)+","+number(b.East))
	}
}

func (q query) extra(p Params) {
	for k, v := range p {
		url.Values(q).Set(k, v)
	}
}

func number(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }

func pair(p Point) string { return number(p.Lat) + "," + number(p.Lon) }

func points(list []Point) string {
	s := make([]string, len(list))
	for i, p := range list {
		s[i] = pair(p)
	}
	return strings.Join(s, "|")
}

func pairs(list []Point) [][2]float64 {
	out := make([][2]float64, len(list))
	for i, p := range list {
		out[i] = [2]float64{p.Lat, p.Lon}
	}
	return out
}

// options adds every field of a params struct that has a q tag and is set: text when not empty, numbers when not
// zero, flags as 1 when true, lists joined with commas, points as lat,lon (several joined with |), boxes as
// south,west,north,east. Fields tagged "-" are left to the method.
func (q query) options(p interface{}) {
	v := reflect.ValueOf(p).Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Tag.Get("q")
		if name == "" || name == "-" {
			continue
		}
		switch f := v.Field(i).Interface().(type) {
		case string:
			q.text(name, f)
		case int:
			q.whole(name, int64(f))
		case int64:
			q.whole(name, f)
		case bool:
			q.flag(name, f)
		case []string:
			q.list(name, f)
		case *Point:
			q.point(name, f)
		case *BBox:
			q.box(name, f)
		case []Point:
			if len(f) > 0 {
				q.text(name, points(f))
			}
		}
	}
}

func formsUsed(locations []Point, locationsText string, path []Point, pathText string, bounds *BBox) bool {
	return len(locations) > 0 || locationsText != "" || len(path) > 0 || pathText != "" || bounds != nil
}

// lang is the lang sent on the endpoints that take it: the call's, else the client's ("en" unless set).
func (c *Client) langOf(call string) string {
	if call != "" {
		return call
	}
	return c.lang
}

// language is timezone's language: the call's, else the client's lang only when the caller set one.
func (c *Client) languageOf(lang, language string) string {
	if lang != "" {
		return lang
	}
	if language != "" {
		return language
	}
	if c.langSet {
		return c.lang
	}
	return ""
}

func (c *Client) get(ctx context.Context, path string, q query, out interface{}) error {
	return c.send(ctx, http.MethodGet, path, q, nil, out)
}

func (c *Client) send(ctx context.Context, method, path string, q query, body, out interface{}) error {
	data, meta, err := c.call(ctx, method, path, url.Values(q), body)
	if err != nil {
		return err
	}
	decodeAnswer(data, meta, out)
	return nil
}

// ---- the 16 endpoint methods -------------------------------------------------------------------------------------

// Forward turns an address or place name into coordinates. Leave query empty to search by the structured fields or
// the PlaceID of p instead.
func (c *Client) Forward(ctx context.Context, query string, p *ForwardParams) (*ForwardResponse, error) {
	if p == nil {
		p = &ForwardParams{}
	}
	q := newQuery()
	q.text("q", query)
	q.options(p)
	q.text("lang", c.langOf(p.Lang))
	q.extra(p.Params)
	out := &ForwardResponse{}
	if err := c.get(ctx, "forward", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Reverse turns a point into the nearest address. Result is nil when nothing is within the radius.
func (c *Client) Reverse(ctx context.Context, lat, lon float64, p *ReverseParams) (*ReverseResponse, error) {
	if p == nil {
		p = &ReverseParams{}
	}
	q := newQuery()
	q.text("lat", number(lat))
	q.text("lon", number(lon))
	c.reverseOptions(q, p)
	out := &ReverseResponse{}
	if err := c.get(ctx, "reverse", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) reverseOptions(q query, p *ReverseParams) {
	q.options(p)
	q.text("lang", c.langOf(p.Lang))
	q.extra(p.Params)
}

// Autocomplete suggests addresses for what someone has typed so far.
func (c *Client) Autocomplete(ctx context.Context, query string, p *AutocompleteParams) (*AutocompleteResponse, error) {
	if p == nil {
		p = &AutocompleteParams{}
	}
	q := newQuery()
	q.text("q", query)
	q.options(p)
	q.text("lang", c.langOf(p.Lang))
	q.extra(p.Params)
	out := &AutocompleteResponse{}
	if err := c.get(ctx, "autocomplete", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ipLookup(ctx context.Context, path, address string, p *IPParams) (*IPInfo, error) {
	if p == nil {
		p = &IPParams{}
	}
	q := newQuery()
	q.text("ip", address)
	q.options(p)
	q.extra(p.Params)
	out := &IPInfo{}
	if err := c.get(ctx, path, q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// IP looks up an IPv4 or IPv6 address. With an empty address it looks up the address the request came from.
func (c *Client) IP(ctx context.Context, address string, p *IPParams) (*IPInfo, error) {
	return c.ipLookup(ctx, "ip", address, p)
}

// IPv4 looks up an IPv4 address. With an empty address it looks up the address the request came from.
func (c *Client) IPv4(ctx context.Context, address string, p *IPParams) (*IPInfo, error) {
	return c.ipLookup(ctx, "ipv4", address, p)
}

// IPv6 looks up an IPv6 address. With an empty address it looks up the address the request came from.
func (c *Client) IPv6(ctx context.Context, address string, p *IPParams) (*IPInfo, error) {
	return c.ipLookup(ctx, "ipv6", address, p)
}

// Timezone returns the zone, offset and local time of a point. With p.Tz it answers for a zone id instead, and with
// p.Locations, p.Path or a p.Bounds grid for several points (TimezoneBatch reads those for you; here they are in
// Raw); lat and lon are then not sent.
func (c *Client) Timezone(ctx context.Context, lat, lon float64, p *TimezoneParams) (*TimezoneInfo, error) {
	if p == nil {
		p = &TimezoneParams{}
	}
	q := newQuery()
	if p.Tz == "" && !formsUsed(p.Locations, p.LocationsText, p.Path, p.PathText, p.Bounds) {
		q.text("lat", number(lat))
		q.text("lon", number(lon))
	}
	q.options(p)
	q.text("language", c.languageOf(p.Lang, p.Language))
	q.extra(p.Params)
	out := &TimezoneInfo{}
	if err := c.get(ctx, "timezone", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Elevation returns the height of a point. With p.Locations, p.Path or a p.Bounds grid it answers for several
// points; lat and lon are then not sent. Results always holds the points, one entry for a single point.
func (c *Client) Elevation(ctx context.Context, lat, lon float64, p *ElevationParams) (*ElevationResponse, error) {
	if p == nil {
		p = &ElevationParams{}
	}
	q := newQuery()
	if !formsUsed(p.Locations, p.LocationsText, p.Path, p.PathText, p.Bounds) {
		q.text("lat", number(lat))
		q.text("lon", number(lon))
	}
	q.options(p)
	q.extra(p.Params)
	out := &ElevationResponse{}
	if err := c.get(ctx, "elevation", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Postcode returns every place that uses a postal code.
func (c *Client) Postcode(ctx context.Context, code string, p *PostcodeParams) (*PostcodeResponse, error) {
	if p == nil {
		p = &PostcodeParams{}
	}
	q := newQuery()
	q.text("code", code)
	q.options(p)
	q.extra(p.Params)
	out := &PostcodeResponse{}
	if err := c.get(ctx, "postcode", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ForwardBatch runs up to 100 searches in one request; each counts as one request. The options apply to every
// search and go in the query string; the body holds only the list.
func (c *Client) ForwardBatch(ctx context.Context, queries []string, p *ForwardBatchParams) (*ForwardBatchResponse, error) {
	if p == nil {
		p = &ForwardBatchParams{}
	}
	if queries == nil {
		queries = []string{}
	}
	q := newQuery()
	q.options(p)
	q.text("lang", c.langOf(p.Lang))
	q.extra(p.Params)
	out := &ForwardBatchResponse{}
	body := map[string]interface{}{"queries": queries}
	if err := c.send(ctx, http.MethodPost, "forward", q, body, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ReverseBatch turns up to 100 points into addresses in one request; each counts as one request.
func (c *Client) ReverseBatch(ctx context.Context, points []Point, p *ReverseParams) (*ReverseBatchResponse, error) {
	if p == nil {
		p = &ReverseParams{}
	}
	q := newQuery()
	c.reverseOptions(q, p)
	out := &ReverseBatchResponse{}
	body := map[string]interface{}{"coords": pairs(points)}
	if err := c.send(ctx, http.MethodPost, "reverse", q, body, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ipBatch(ctx context.Context, path string, addresses []string, p *IPParams) (*IPBatchResponse, error) {
	if p == nil {
		p = &IPParams{}
	}
	q := newQuery()
	url.Values(q).Set("ips", strings.Join(addresses, ","))
	q.options(p)
	q.extra(p.Params)
	out := &IPBatchResponse{}
	if err := c.get(ctx, path, q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// IPBatch looks up to 100 addresses of either version in one request.
func (c *Client) IPBatch(ctx context.Context, addresses []string, p *IPParams) (*IPBatchResponse, error) {
	return c.ipBatch(ctx, "ip", addresses, p)
}

// IPv4Batch looks up to 100 IPv4 addresses in one request.
func (c *Client) IPv4Batch(ctx context.Context, addresses []string, p *IPParams) (*IPBatchResponse, error) {
	return c.ipBatch(ctx, "ipv4", addresses, p)
}

// IPv6Batch looks up to 100 IPv6 addresses in one request.
func (c *Client) IPv6Batch(ctx context.Context, addresses []string, p *IPParams) (*IPBatchResponse, error) {
	return c.ipBatch(ctx, "ipv6", addresses, p)
}

// TimezoneBatch returns the timezones of up to 100 points in one request. With p.LocationsText the points argument
// is not used (it may be nil).
func (c *Client) TimezoneBatch(ctx context.Context, list []Point, p *TimezoneBatchParams) (*TimezoneBatchResponse, error) {
	if p == nil {
		p = &TimezoneBatchParams{}
	}
	q := newQuery()
	if p.LocationsText != "" {
		url.Values(q).Set("locations", p.LocationsText)
	} else {
		url.Values(q).Set("locations", points(list))
	}
	q.options(p)
	q.text("language", c.languageOf(p.Lang, p.Language))
	q.extra(p.Params)
	out := &TimezoneBatchResponse{}
	if err := c.get(ctx, "timezone", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ElevationBatch returns the heights of up to 100 points, sent in a POST body.
func (c *Client) ElevationBatch(ctx context.Context, points []Point, p *ElevationBatchParams) (*ElevationResponse, error) {
	if p == nil {
		p = &ElevationBatchParams{}
	}
	q := newQuery()
	q.options(p)
	q.extra(p.Params)
	out := &ElevationResponse{}
	body := map[string]interface{}{"locations": pairs(points)}
	if err := c.send(ctx, http.MethodPost, "elevation", q, body, out); err != nil {
		return nil, err
	}
	return out, nil
}

func newQuery() query { return query(url.Values{}) }
