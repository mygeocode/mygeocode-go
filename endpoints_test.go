package mygeocode

import (
	"encoding/json"
	"reflect"
	"testing"
)

// Every method sends the shared example values; the request must equal the recorded one and the answer must be
// read into the right types (the recordings in testdata are real answers of the API).

func TestForward(t *testing.T) {
	rec := newRecorder(replay(t, "forward"))
	s := &setup{}
	r, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil)
	if err != nil {
		t.Fatal(err)
	}
	sentLike(t, "forward", rec.last())
	want := bodyOf(t, "forward")["results"].([]interface{})[0].(map[string]interface{})
	str(t, r.Status, "ok")
	str(t, r.Results[0].Formatted, "Brandenburg Gate, State of Berlin, Germany")
	num(t, r.Results[0].Lat, want["lat"].(float64))
	num(t, r.Results[0].Lon, want["lon"].(float64))
	str(t, r.Results[0].Components.CountryCode, want["components"].(map[string]interface{})["country_code"].(string))
	if string(r.Raw) != string(load(t, "forward").Body) {
		t.Fatal("Raw is not the answer")
	}
}

func TestForwardMetaComesFromTheHeaders(t *testing.T) {
	rec := newRecorder(replay(t, "forward"))
	s := &setup{}
	r, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", nil)
	if err != nil {
		t.Fatal(err)
	}
	h := load(t, "forward").Headers
	if r.Meta.RequestID != h["x-request-id"] {
		t.Fatal(r.Meta.RequestID)
	}
	for _, c := range []struct {
		got  *int64
		name string
	}{{r.Meta.QuotaLimit, "x-quota-limit"}, {r.Meta.QuotaUsed, "x-quota-used"}, {r.Meta.QuotaFreeRemaining, "x-quota-free-remaining"}, {r.Meta.QuotaReset, "x-quota-reset"}} {
		if c.got == nil || !same(h[c.name], toText(float64(*c.got))) {
			t.Fatalf("%s: %v vs %s", c.name, show(c.got), h[c.name])
		}
	}
	if r.Meta.CreditsRemaining != nil {
		t.Fatal("CreditsRemaining should be nil")
	}
}

func TestForwardStructured(t *testing.T) {
	rec := newRecorder(replay(t, "forward_structured"))
	s := &setup{}
	r, err := s.client(rec).Forward(bg, "", &ForwardParams{Street: "10 Downing Street", City: "London", Country: "GB"})
	if err != nil {
		t.Fatal(err)
	}
	sentLike(t, "forward_structured", rec.last())
	str(t, r.Results[0].Formatted, "Downing Street, London SW1A, United Kingdom")
}

func TestReverse(t *testing.T) {
	rec := newRecorder(replay(t, "reverse"))
	s := &setup{}
	r, err := s.client(rec).Reverse(bg, eiffel.Lat, eiffel.Lon, nil)
	if err != nil {
		t.Fatal(err)
	}
	sentLike(t, "reverse", rec.last())
	str(t, r.Result.Formatted, "5 Avenue Anatole France, 75007 Paris, France")
	whole(t, r.Result.DistanceM, 47)
}

func TestAutocomplete(t *testing.T) {
	rec := newRecorder(replay(t, "autocomplete"))
	s := &setup{}
	r, err := s.client(rec).Autocomplete(bg, "1600 Pennsylvania", nil)
	if err != nil {
		t.Fatal(err)
	}
	sentLike(t, "autocomplete", rec.last())
	str(t, r.Suggestions[0].Text, "1600 Pennsylvania Avenue NW, Washington, DC 20500, United States")
}

func TestIPIPv4IPv6(t *testing.T) {
	s := &setup{}
	for _, c := range []struct {
		name, address string
	}{{"ip", "8.8.8.8"}, {"ipv4", "8.8.8.8"}, {"ipv6", "2001:4860:4860::8888"}} {
		rec := newRecorder(replay(t, c.name))
		cl := s.client(rec)
		call := map[string]func(string, *IPParams) (*IPInfo, error){
			"ip":   func(a string, p *IPParams) (*IPInfo, error) { return cl.IP(bg, a, p) },
			"ipv4": func(a string, p *IPParams) (*IPInfo, error) { return cl.IPv4(bg, a, p) },
			"ipv6": func(a string, p *IPParams) (*IPInfo, error) { return cl.IPv6(bg, a, p) },
		}[c.name]
		r, err := call(c.address, nil)
		if err != nil {
			t.Fatal(err)
		}
		sentLike(t, c.name, rec.last())
		str(t, r.IP, c.address)
		str(t, r.ASName, "Google LLC")
	}
	rec := newRecorder(replay(t, "ip"))
	r, err := s.client(rec).IP(bg, "8.8.8.8", nil)
	if err != nil {
		t.Fatal(err)
	}
	str(t, r.Country, "United States")
	str(t, r.City, "Mountain View")
	whole(t, r.ASN, 15169)
	if !BoolValue(r.IsDatacenter) {
		t.Fatal("IsDatacenter")
	}
}

func TestIPOfTheCallerSendsNoAddress(t *testing.T) {
	rec := newRecorder(replay(t, "ip"))
	s := &setup{}
	if _, err := s.client(rec).IP(bg, "", nil); err != nil {
		t.Fatal(err)
	}
	if _, has := rec.last().query["ip"]; has || rec.last().url.Path != "/v1/ip" {
		t.Fatal(rec.last().url)
	}
}

func TestTimezone(t *testing.T) {
	rec := newRecorder(replay(t, "timezone"))
	s := &setup{}
	r, err := s.client(rec).Timezone(bg, 35.6762, 139.6503, nil)
	if err != nil {
		t.Fatal(err)
	}
	sentLike(t, "timezone", rec.last())
	str(t, r.Timezone, "Asia/Tokyo")
	str(t, r.UTCOffset, "+09:00")
	str(t, r.Name, "Japan Standard Time")
	str(t, r.Abbreviation, "JST")
}

func TestTimezoneByZoneID(t *testing.T) {
	rec := newRecorder(replay(t, "timezone_tz"))
	s := &setup{}
	r, err := s.client(rec).Timezone(bg, 0, 0, &TimezoneParams{Tz: "Europe/Paris"})
	if err != nil {
		t.Fatal(err)
	}
	sentLike(t, "timezone_tz", rec.last())
	str(t, r.Timezone, "Europe/Paris")
}

func TestElevation(t *testing.T) {
	rec := newRecorder(replay(t, "elevation"))
	s := &setup{}
	r, err := s.client(rec).Elevation(bg, 27.9881, 86.9250, nil)
	if err != nil {
		t.Fatal(err)
	}
	sentLike(t, "elevation", rec.last())
	num(t, r.Results[0].ElevationM, 8627)
}

func TestElevationAnsweredFlatIsReadAsAListOfOne(t *testing.T) {
	flat := map[string]interface{}{"status": "ok", "lat": 27.9881, "lon": 86.925, "elevation_m": 8627, "resolution_m": 30}
	rec := newRecorder(respond(200, flat, nil))
	s := &setup{}
	r, err := s.client(rec).Elevation(bg, 27.9881, 86.9250, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Results) != 1 {
		t.Fatal(len(r.Results))
	}
	num(t, r.Results[0].ElevationM, 8627)
	num(t, r.Results[0].ResolutionM, 30)
}

func TestPostcode(t *testing.T) {
	rec := newRecorder(replay(t, "postcode"))
	s := &setup{}
	r, err := s.client(rec).Postcode(bg, "10001", &PostcodeParams{Country: "US"})
	if err != nil {
		t.Fatal(err)
	}
	sentLike(t, "postcode", rec.last())
	str(t, r.Results[0].Place, "New York")
	str(t, r.Results[0].RegionCode, "NY")
}

func TestForwardBatch(t *testing.T) {
	rec := newRecorder(replay(t, "forward_batch"))
	s := &setup{}
	r, err := s.client(rec).ForwardBatch(bg, []string{"Brandenburg Gate, Berlin", "Sydney Opera House", "Statue of Liberty, New York"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	sentLike(t, "forward_batch", rec.last())
	want := [][2]string{
		{"Brandenburg Gate, Berlin", "Brandenburg Gate, State of Berlin, Germany"},
		{"Sydney Opera House", "Sydney Opera House, Sydney, New South Wales, Australia"},
		{"Statue of Liberty, New York", "Statue of Liberty, New York, New York, United States"},
	}
	for i, w := range want {
		str(t, r.Results[i].Query, w[0])
		str(t, r.Results[i].Results[0].Formatted, w[1])
	}
}

func TestForwardBatchReadsANumericPostcodeAsText(t *testing.T) {
	rec := newRecorder(replay(t, "forward_batch"))
	s := &setup{}
	r, err := s.client(rec).ForwardBatch(bg, []string{"Brandenburg Gate, Berlin", "Sydney Opera House", "Statue of Liberty, New York"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	york := r.Results[2].Results[3]
	var raw struct {
		Components struct {
			Postcode interface{} `json:"postcode"`
		} `json:"components"`
	}
	_ = json.Unmarshal(york.Raw, &raw)
	if raw.Components.Postcode != float64(68467) {
		t.Fatalf("raw postcode %v", raw.Components.Postcode)
	}
	str(t, york.Components.Postcode, "68467")
}

func TestReverseBatch(t *testing.T) {
	rec := newRecorder(replay(t, "reverse_batch"))
	s := &setup{}
	r, err := s.client(rec).ReverseBatch(bg, []Point{eiffel, liberty, opera}, nil)
	if err != nil {
		t.Fatal(err)
	}
	sentLike(t, "reverse_batch", rec.last())
	want := []string{
		"5 Avenue Anatole France, 75007 Paris, France",
		"1 Liberty Island, New York, NY 10004, United States",
		"2B Macquarie Street, Sydney NSW 2000, Australia",
	}
	for i, w := range want {
		str(t, r.Results[i].Result.Formatted, w)
	}
}

func TestIPBatches(t *testing.T) {
	s := &setup{}
	cases := []struct {
		name      string
		addresses []string
	}{
		{"ip_batch", []string{"8.8.8.8", "2606:4700:4700::1111"}},
		{"ipv4_batch", []string{"8.8.8.8", "1.1.1.1", "9.9.9.9"}},
		{"ipv6_batch", []string{"2001:4860:4860::8888", "2606:4700:4700::1111"}},
	}
	for _, c := range cases {
		rec := newRecorder(replay(t, c.name))
		cl := s.client(rec)
		var r *IPBatchResponse
		var err error
		switch c.name {
		case "ip_batch":
			r, err = cl.IPBatch(bg, c.addresses, nil)
		case "ipv4_batch":
			r, err = cl.IPv4Batch(bg, c.addresses, nil)
		default:
			r, err = cl.IPv6Batch(bg, c.addresses, nil)
		}
		if err != nil {
			t.Fatal(err)
		}
		sentLike(t, c.name, rec.last())
		for i, a := range c.addresses {
			str(t, r.Results[i].IP, a)
		}
	}
	rec := newRecorder(replay(t, "ipv4_batch"))
	r, err := s.client(rec).IPv4Batch(bg, []string{"8.8.8.8", "1.1.1.1", "9.9.9.9"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i, w := range []string{"Google LLC", "Cloudflare, Inc.", "Quad9"} {
		str(t, r.Results[i].ASName, w)
	}
}

func TestIPBatchEntryTheAPIRejects(t *testing.T) {
	var good interface{}
	_ = json.Unmarshal(load(t, "ipv4").Body, &good)
	body := map[string]interface{}{"status": "ok", "results": []interface{}{good, map[string]interface{}{
		"status": "error", "ip": "not-an-ip", "error": map[string]interface{}{"code": "invalid_ip", "message": "Not an IP address."},
	}}}
	rec := newRecorder(respond(200, body, nil))
	s := &setup{}
	r, err := s.client(rec).IPBatch(bg, []string{"8.8.8.8", "not-an-ip"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	str(t, r.Results[0].Status, "ok")
	str(t, r.Results[1].Status, "error")
	if r.Results[0].Error != nil || r.Results[1].Error == nil {
		t.Fatal("Error")
	}
	str(t, r.Results[1].Error.Code, "invalid_ip")
	str(t, r.Results[1].Error.Message, "Not an IP address.")
}

func TestTimezoneBatch(t *testing.T) {
	rec := newRecorder(replay(t, "timezone_batch"))
	s := &setup{}
	r, err := s.client(rec).TimezoneBatch(bg, []Point{eiffel, liberty, opera}, nil)
	if err != nil {
		t.Fatal(err)
	}
	sentLike(t, "timezone_batch", rec.last())
	for i, w := range []string{"Europe/Paris", "America/New_York", "Australia/Sydney"} {
		str(t, r.Results[i].Timezone, w)
	}
	whole(t, r.Timestamp, int64(bodyOf(t, "timezone_batch")["timestamp"].(float64)))
}

func TestElevationBatch(t *testing.T) {
	rec := newRecorder(replay(t, "elevation_batch"))
	s := &setup{}
	r, err := s.client(rec).ElevationBatch(bg, []Point{eiffel, liberty, opera}, nil)
	if err != nil {
		t.Fatal(err)
	}
	sentLike(t, "elevation_batch", rec.last())
	for i, w := range []float64{37, 0, 10} {
		num(t, r.Results[i].ElevationM, w)
	}
}

func TestUnknownFieldsDoNotBreakParsing(t *testing.T) {
	body := bodyOf(t, "reverse")
	body["brand_new_field"] = map[string]interface{}{"x": 1}
	body["result"].(map[string]interface{})["another_new_field"] = []interface{}{1, 2}
	rec := newRecorder(respond(200, body, nil))
	s := &setup{}
	r, err := s.client(rec).Reverse(bg, eiffel.Lat, eiffel.Lon, nil)
	if err != nil {
		t.Fatal(err)
	}
	var top, inner map[string]interface{}
	_ = json.Unmarshal(r.Raw, &top)
	_ = json.Unmarshal(r.Result.Raw, &inner)
	if !reflect.DeepEqual(top["brand_new_field"], map[string]interface{}{"x": float64(1)}) {
		t.Fatal(top["brand_new_field"])
	}
	if !reflect.DeepEqual(inner["another_new_field"], []interface{}{float64(1), float64(2)}) {
		t.Fatal(inner["another_new_field"])
	}
	str(t, r.Result.Formatted, "5 Avenue Anatole France, 75007 Paris, France")
}

func TestWrongTypesReadAsNil(t *testing.T) {
	body := map[string]interface{}{"status": "ok", "lat": "not a number", "lon": true, "result": map[string]interface{}{
		"formatted": 12, "lat": "x", "distance_m": 4.5, "components": "oops", "codes": nil, "bounds": []interface{}{1},
	}}
	rec := newRecorder(respond(200, body, nil))
	s := &setup{}
	r, err := s.client(rec).Reverse(bg, eiffel.Lat, eiffel.Lon, nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.Lat != nil || r.Lon != nil || r.Result.Lat != nil || r.Result.DistanceM != nil || r.Result.Components != nil || r.Result.Codes != nil || r.Result.Bounds != nil {
		t.Fatalf("%+v", r.Result)
	}
	str(t, r.Result.Formatted, "12")
}
