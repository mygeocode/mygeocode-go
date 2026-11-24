package mygeocode

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

// The answers are read leniently, field by field, so that one field of an unexpected type never fails a whole
// call: a field that is missing, null or of the wrong type is left nil, a number in a text field becomes text, a
// list that is missing or null is empty, and the raw JSON of every object is kept in its Raw field.

var (
	rawType     = reflect.TypeOf(json.RawMessage(nil))
	stringsType = reflect.TypeOf([]string(nil))
	metaType    = reflect.TypeOf(Meta{})
	// pointerKind is the kind of a pointer, taken from a value: reflect.Ptr is its old name and reflect.Pointer
	// only exists from Go 1.18.
	pointerKind = reflect.TypeOf((*int)(nil)).Kind()
)

// objectOf reads data as a JSON object, keeping each value's own bytes.
func objectOf(data []byte) (map[string]json.RawMessage, bool) {
	if kindOf(data) != '{' {
		return nil, false
	}
	var obj map[string]json.RawMessage
	if json.Unmarshal(data, &obj) != nil {
		return nil, false
	}
	return obj, true
}

// kindOf is the first byte of a JSON value: '{', '[', '"', 't', 'f', 'n', or '0' for a number; 0 when empty.
func kindOf(data []byte) byte {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return 0
	}
	switch c := data[0]; {
	case c == '-' || (c >= '0' && c <= '9'):
		return '0'
	default:
		return c
	}
}

// textOf reads a JSON string, or a number as its text; anything else is "".
func textOf(data []byte) string {
	if s := textPtr(data); s != nil {
		return *s
	}
	return ""
}

func textPtr(data []byte) *string {
	switch kindOf(data) {
	case '"':
		var s string
		if json.Unmarshal(data, &s) == nil {
			return &s
		}
	case '0':
		var n json.Number
		if json.Unmarshal(data, &n) == nil {
			s := n.String()
			return &s
		}
	}
	return nil
}

func floatPtr(data []byte) *float64 {
	if kindOf(data) != '0' {
		return nil
	}
	var f float64
	if json.Unmarshal(data, &f) != nil {
		return nil
	}
	return &f
}

func intPtr(data []byte) *int64 {
	if kindOf(data) != '0' {
		return nil
	}
	var n json.Number
	if json.Unmarshal(data, &n) != nil {
		return nil
	}
	if i, err := strconv.ParseInt(n.String(), 10, 64); err == nil {
		return &i
	}
	f, err := n.Float64()
	if err != nil || f != float64(int64(f)) {
		return nil
	}
	i := int64(f)
	return &i
}

func boolPtr(data []byte) *bool {
	k := kindOf(data)
	if k != 't' && k != 'f' {
		return nil
	}
	var b bool
	if json.Unmarshal(data, &b) != nil {
		return nil
	}
	return &b
}

// decodeInto fills the struct v points to from the JSON object obj, whose own bytes are raw.
func decodeInto(v reflect.Value, obj map[string]json.RawMessage, raw []byte) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)
		if f.Name == "Raw" && f.Type == rawType {
			fv.Set(reflect.ValueOf(copyBytes(raw)))
			continue
		}
		name := strings.Split(f.Tag.Get("json"), ",")[0]
		if name == "" || name == "-" || f.Type == metaType {
			continue
		}
		data, ok := obj[name]
		if !ok {
			if f.Type.Kind() == reflect.Slice && f.Type != rawType {
				fv.Set(reflect.MakeSlice(f.Type, 0, 0))
			}
			continue
		}
		decodeField(fv, data)
	}
}

func decodeField(fv reflect.Value, data []byte) {
	t := fv.Type()
	switch {
	case t == rawType:
		if k := kindOf(data); k != 0 && k != 'n' {
			fv.Set(reflect.ValueOf(json.RawMessage(copyBytes(bytes.TrimSpace(data)))))
		}
	case t == stringsType:
		var items []json.RawMessage
		out := make([]string, 0)
		if kindOf(data) == '[' && json.Unmarshal(data, &items) == nil {
			for _, item := range items {
				out = append(out, textOf(item))
			}
		}
		fv.Set(reflect.ValueOf(out))
	case t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Struct:
		var items []json.RawMessage
		out := reflect.MakeSlice(t, 0, 0)
		if kindOf(data) == '[' && json.Unmarshal(data, &items) == nil {
			for _, item := range items {
				entry := reflect.New(t.Elem()).Elem()
				if obj, ok := objectOf(item); ok {
					decodeInto(entry, obj, item)
				} else {
					decodeInto(entry, map[string]json.RawMessage{}, nil)
				}
				out = reflect.Append(out, entry)
			}
		}
		fv.Set(out)
	case t.Kind() == pointerKind:
		switch t.Elem().Kind() {
		case reflect.String:
			if p := textPtr(data); p != nil {
				fv.Set(reflect.ValueOf(p))
			}
		case reflect.Float64:
			if p := floatPtr(data); p != nil {
				fv.Set(reflect.ValueOf(p))
			}
		case reflect.Int64:
			if p := intPtr(data); p != nil {
				fv.Set(reflect.ValueOf(p))
			}
		case reflect.Bool:
			if p := boolPtr(data); p != nil {
				fv.Set(reflect.ValueOf(p))
			}
		case reflect.Struct:
			if obj, ok := objectOf(data); ok {
				p := reflect.New(t.Elem())
				decodeInto(p.Elem(), obj, data)
				fv.Set(p)
			}
		}
	}
}

func copyBytes(b []byte) []byte {
	if b == nil {
		return nil
	}
	out := make([]byte, len(b))
	copy(out, b)
	return out
}

// decodeAnswer fills out (a pointer to a result struct) from a JSON object answer and sets its Meta.
func decodeAnswer(body []byte, meta Meta, out interface{}) {
	obj, _ := objectOf(body)
	if obj == nil {
		obj = map[string]json.RawMessage{}
	}
	if _, isElevation := out.(*ElevationResponse); isElevation {
		if _, has := obj["results"]; !has {
			if _, flat := obj["elevation_m"]; flat {
				// A single point answered flat: read it as a list of one.
				obj["results"] = json.RawMessage("[" + string(bytes.TrimSpace(body)) + "]")
			}
		}
	}
	v := reflect.ValueOf(out).Elem()
	decodeInto(v, obj, body)
	if m := v.FieldByName("Meta"); m.IsValid() && m.Type() == metaType {
		m.Set(reflect.ValueOf(meta))
	}
}

// StringValue returns *p, or "" when p is nil.
func StringValue(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// FloatValue returns *p, or 0 when p is nil.
func FloatValue(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

// IntValue returns *p, or 0 when p is nil.
func IntValue(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

// BoolValue returns *p, or false when p is nil.
func BoolValue(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}
