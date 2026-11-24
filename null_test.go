package mygeocode

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

// The API may one day answer a field with null instead of dropping it (fields this library reads are never removed;
// they become null). Every method must keep working: a null field stays nil (never a made-up 0 or ""), a null list
// reads as empty and a null object as nil, with no error.
//
// Each recorded answer is replayed twice: once with every scalar set to null ("leaves"), once with every value set to
// null, lists and objects included ("all"). Only the top-level "status" stays "ok", so the answer is not an error.

type nullCall struct {
	name string
	call func(c *Client) (interface{}, error)
}

func nullCalls() []nullCall {
	return []nullCall{
		{"forward", func(c *Client) (interface{}, error) { return c.Forward(bg, "Brandenburg Gate, Berlin", nil) }},
		{"reverse", func(c *Client) (interface{}, error) { return c.Reverse(bg, eiffel.Lat, eiffel.Lon, nil) }},
		{"autocomplete", func(c *Client) (interface{}, error) { return c.Autocomplete(bg, "1600 Pennsylvania", nil) }},
		{"ip", func(c *Client) (interface{}, error) { return c.IP(bg, "8.8.8.8", nil) }},
		{"ipv4", func(c *Client) (interface{}, error) { return c.IPv4(bg, "8.8.8.8", nil) }},
		{"ipv6", func(c *Client) (interface{}, error) { return c.IPv6(bg, "2001:4860:4860::8888", nil) }},
		{"timezone", func(c *Client) (interface{}, error) { return c.Timezone(bg, 35.6762, 139.6503, nil) }},
		{"elevation", func(c *Client) (interface{}, error) { return c.Elevation(bg, 27.9881, 86.9250, nil) }},
		{"postcode", func(c *Client) (interface{}, error) { return c.Postcode(bg, "10001", &PostcodeParams{Country: "US"}) }},
		{"forward_batch", func(c *Client) (interface{}, error) {
			return c.ForwardBatch(bg, []string{"Brandenburg Gate, Berlin", "Sydney Opera House"}, nil)
		}},
		{"reverse_batch", func(c *Client) (interface{}, error) { return c.ReverseBatch(bg, []Point{eiffel, liberty}, nil) }},
		{"ip_batch", func(c *Client) (interface{}, error) {
			return c.IPBatch(bg, []string{"8.8.8.8", "2606:4700:4700::1111"}, nil)
		}},
		{"ipv4_batch", func(c *Client) (interface{}, error) {
			return c.IPv4Batch(bg, []string{"8.8.8.8", "1.1.1.1", "9.9.9.9"}, nil)
		}},
		{"ipv6_batch", func(c *Client) (interface{}, error) { return c.IPv6Batch(bg, []string{"2001:4860:4860::8888"}, nil) }},
		{"timezone_batch", func(c *Client) (interface{}, error) { return c.TimezoneBatch(bg, []Point{eiffel, liberty}, nil) }},
		{"elevation_batch", func(c *Client) (interface{}, error) { return c.ElevationBatch(bg, []Point{eiffel, liberty}, nil) }},
	}
}

func nullify(v interface{}, everything, top bool) interface{} {
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	out := map[string]interface{}{}
	for k, item := range m {
		switch x := item.(type) {
		case map[string]interface{}:
			if top && k == "status" {
				out[k] = "ok"
			} else if everything {
				out[k] = nil
			} else {
				out[k] = nullify(x, false, false)
			}
		case []interface{}:
			if everything {
				out[k] = nil
				continue
			}
			list := make([]interface{}, len(x))
			for i, e := range x {
				if _, isMap := e.(map[string]interface{}); isMap {
					list[i] = nullify(e, false, false)
				}
			}
			out[k] = list
		default:
			if top && k == "status" {
				out[k] = "ok"
			} else {
				out[k] = nil
			}
		}
	}
	return out
}

// allNull is true when every leaf of a JSON value is null.
func allNull(data []byte) bool {
	var v interface{}
	if json.Unmarshal(data, &v) != nil {
		return false
	}
	var walk func(interface{}) bool
	walk = func(v interface{}) bool {
		switch x := v.(type) {
		case nil:
			return true
		case map[string]interface{}:
			for _, e := range x {
				if !walk(e) {
					return false
				}
			}
			return true
		case []interface{}:
			for _, e := range x {
				if !walk(e) {
					return false
				}
			}
			return true
		}
		return false
	}
	return walk(v)
}

// nonNull lists every field of the result tree (Meta and Raw aside) that is not nil, empty, or only nulls.
func nonNull(v reflect.Value, path string, found *[]string) {
	switch v.Kind() {
	case pointerKind:
		if v.IsNil() {
			return
		}
		if v.Elem().Kind() == reflect.Struct {
			nonNull(v.Elem(), path, found)
			return
		}
		*found = append(*found, fmt.Sprintf("%s = %v", path, v.Elem().Interface()))
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			name := t.Field(i).Name
			if name == "Meta" || name == "Raw" {
				continue
			}
			f := v.Field(i)
			if name == "Status" && path == "" {
				continue
			}
			nonNull(f, path+"."+name, found)
		}
	case reflect.Slice:
		if v.Type() == rawType {
			if v.Len() > 0 && !allNull(v.Bytes()) {
				*found = append(*found, fmt.Sprintf("%s = %s", path, v.Bytes()))
			}
			return
		}
		for i := 0; i < v.Len(); i++ {
			e := v.Index(i)
			if e.Kind() == reflect.String {
				if e.String() != "" {
					*found = append(*found, fmt.Sprintf("%s[%d] = %q", path, i, e.String()))
				}
				continue
			}
			nonNull(e, fmt.Sprintf("%s[%d]", path, i), found)
		}
	}
}

func checkNulls(t *testing.T, everything bool) {
	for _, c := range nullCalls() {
		var body interface{}
		_ = json.Unmarshal(load(t, c.name).Body, &body)
		rec := newRecorder(respond(200, nullify(body, everything, true), nil))
		s := &setup{}
		result, err := c.call(s.client(rec))
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		var found []string
		nonNull(reflect.ValueOf(result), "", &found)
		if len(found) > 0 {
			t.Fatalf("%s (everything=%v): fields must stay nil: %v", c.name, everything, found)
		}
		status := reflect.ValueOf(result).Elem().FieldByName("Status").Interface().(*string)
		str(t, status, "ok")
	}
}

func TestEveryScalarNullStaysNil(t *testing.T) { checkNulls(t, false) }

func TestEveryValueNullIncludingListsAndObjects(t *testing.T) { checkNulls(t, true) }

func TestNullListsReadAsEmptyLists(t *testing.T) {
	s := &setup{}
	f, err := s.client(newRecorder(respond(200, map[string]interface{}{"status": "ok", "query": nil, "results": nil}, nil))).Forward(bg, "Brandenburg Gate, Berlin", nil)
	if err != nil {
		t.Fatal(err)
	}
	if f.Results == nil || len(f.Results) != 0 {
		t.Fatal("results should be an empty list")
	}
	r, err := s.client(newRecorder(respond(200, map[string]interface{}{"status": "ok", "lat": nil, "lon": nil, "result": nil}, nil))).Reverse(bg, eiffel.Lat, eiffel.Lon, nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.Result != nil || r.Lat != nil {
		t.Fatal("result and lat should be nil")
	}
	ip, err := s.client(newRecorder(respond(200, map[string]interface{}{"status": "ok", "ip": "8.8.8.8", "tags": nil}, nil))).IP(bg, "8.8.8.8", nil)
	if err != nil {
		t.Fatal(err)
	}
	if ip.Tags == nil || len(ip.Tags) != 0 {
		t.Fatal("tags should be an empty list")
	}
	e, err := s.client(newRecorder(respond(200, map[string]interface{}{"status": "ok"}, nil))).ElevationBatch(bg, []Point{eiffel}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if e.Results == nil {
		t.Fatal("missing results should be an empty list")
	}
}
