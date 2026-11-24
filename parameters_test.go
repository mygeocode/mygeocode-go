package mygeocode

import (
	"reflect"
	"testing"
)

// Every documented parameter is a field of its method's params struct, sent under the API's own name.

func TestForwardEveryParameter(t *testing.T) {
	rec := newRecorder(replay(t, "forward"))
	s := &setup{}
	_, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", &ForwardParams{
		Street: "Pariser Platz", City: "Berlin", State: "Berlin", Postcode: "10117", Country: "DE", PlaceID: "abc",
		Limit: 3, Countries: []string{"de", "at"}, Bounds: &BBox{52.3, 13.0, 52.7, 13.8}, Strict: true,
		Proximity: &Point{52.5, 13.4}, Types: []string{"poi", "address"}, Lang: "de", Annotations: true, Pretty: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	eqQuery(t, rec.last().query, map[string]string{
		"q": "Brandenburg Gate, Berlin", "street": "Pariser Platz", "city": "Berlin", "state": "Berlin",
		"postcode": "10117", "country": "DE", "place_id": "abc", "limit": "3", "countries": "de,at",
		"bounds": "52.3,13,52.7,13.8", "strict": "1", "proximity": "52.5,13.4", "types": "poi,address",
		"lang": "de", "annotations": "1", "pretty": "1",
	})
}

var flagFields = map[string]string{
	"timezone": "Timezone", "elevation": "Elevation", "wikidata": "Wikidata", "country_info": "CountryInfo",
	"codes": "Codes", "sun": "Sun", "regions": "Regions", "districts": "Districts", "census": "Census",
	"demographics": "Demographics", "road": "Road", "site": "Site", "annotations": "Annotations",
}

func withFlag(p interface{}, field string) {
	reflect.ValueOf(p).Elem().FieldByName(field).SetBool(true)
}

func TestFlagsOnForwardReverseAndTheirBatches(t *testing.T) {
	s := &setup{}
	for name, field := range flagFields {
		rec := newRecorder(replay(t, "forward"))
		c := s.client(rec)
		fp, rp, bp, rbp := &ForwardParams{}, &ReverseParams{}, &ForwardBatchParams{}, &ReverseParams{}
		withFlag(fp, field)
		withFlag(rp, field)
		withFlag(bp, field)
		withFlag(rbp, field)
		if _, err := c.Forward(bg, "Brandenburg Gate, Berlin", fp); err != nil {
			t.Fatal(err)
		}
		if rec.last().query[name] != "1" {
			t.Fatalf("forward %s", name)
		}
		if _, err := c.Reverse(bg, eiffel.Lat, eiffel.Lon, rp); err != nil {
			t.Fatal(err)
		}
		if rec.last().query[name] != "1" {
			t.Fatalf("reverse %s", name)
		}
		if _, err := c.ForwardBatch(bg, []string{"Brandenburg Gate, Berlin"}, bp); err != nil {
			t.Fatal(err)
		}
		if rec.last().query[name] != "1" {
			t.Fatalf("forward batch %s", name)
		}
		if _, err := c.ReverseBatch(bg, []Point{eiffel}, rbp); err != nil {
			t.Fatal(err)
		}
		if rec.last().query[name] != "1" {
			t.Fatalf("reverse batch %s", name)
		}
	}
	// Every params struct with flags has exactly these, under these names.
	for _, p := range []interface{}{ForwardParams{}, ForwardBatchParams{}, ReverseParams{}} {
		tp := reflect.TypeOf(p)
		for name, field := range flagFields {
			f, ok := tp.FieldByName(field)
			if !ok || f.Tag.Get("q") != name {
				t.Fatalf("%s.%s", tp.Name(), field)
			}
		}
	}
}

func TestReverseEveryParameter(t *testing.T) {
	rec := newRecorder(replay(t, "reverse"))
	s := &setup{}
	_, err := s.client(rec).Reverse(bg, eiffel.Lat, eiffel.Lon, &ReverseParams{
		Level: "city", Radius: 500, Lang: "fr", Boundaries: true, Annotations: true, Pretty: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	eqQuery(t, rec.last().query, map[string]string{
		"lat": "48.8584", "lon": "2.2945", "level": "city", "radius": "500", "lang": "fr", "boundaries": "1",
		"annotations": "1", "pretty": "1",
	})
}

func TestAutocompleteEveryParameter(t *testing.T) {
	rec := newRecorder(replay(t, "autocomplete"))
	s := &setup{}
	_, err := s.client(rec).Autocomplete(bg, "1600 Pennsylvania", &AutocompleteParams{
		Limit: 5, Countries: []string{"us"}, Proximity: &Point{38.9, -77.0}, Bounds: &BBox{38, -78, 39, -76},
		Strict: true, Types: []string{"address"}, Lang: "en", Pretty: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	eqQuery(t, rec.last().query, map[string]string{
		"q": "1600 Pennsylvania", "limit": "5", "countries": "us", "proximity": "38.9,-77",
		"bounds": "38,-78,39,-76", "strict": "1", "types": "address", "lang": "en", "pretty": "1",
	})
}

func TestIPParameters(t *testing.T) {
	rec := newRecorder(replay(t, "ip"))
	s := &setup{}
	c := s.client(rec)
	if _, err := c.IPv4(bg, "8.8.8.8", &IPParams{Elevation: true, Pretty: true}); err != nil {
		t.Fatal(err)
	}
	eqQuery(t, rec.last().query, map[string]string{"ip": "8.8.8.8", "elevation": "1", "pretty": "1"})
	if _, err := c.IPv6Batch(bg, []string{"2001:4860:4860::8888", "2606:4700:4700::1111"}, &IPParams{Elevation: true}); err != nil {
		t.Fatal(err)
	}
	eqQuery(t, rec.last().query, map[string]string{"ips": "2001:4860:4860::8888,2606:4700:4700::1111", "elevation": "1"})
}

func TestTimezoneParameters(t *testing.T) {
	rec := newRecorder(replay(t, "timezone"))
	s := &setup{}
	c := s.client(rec)
	run := func(lat, lon float64, p *TimezoneParams, want map[string]string) {
		t.Helper()
		if _, err := c.Timezone(bg, lat, lon, p); err != nil {
			t.Fatal(err)
		}
		eqQuery(t, rec.last().query, want)
	}
	run(35.6762, 139.6503, &TimezoneParams{Timestamp: 1767225600, Lang: "ja", Pretty: true},
		map[string]string{"lat": "35.6762", "lon": "139.6503", "timestamp": "1767225600", "language": "ja", "pretty": "1"})
	run(0, 0, &TimezoneParams{Locations: []Point{eiffel, liberty}},
		map[string]string{"locations": "48.8584,2.2945|40.6892,-74.0445"})
	run(0, 0, &TimezoneParams{LocationsText: "enc:abc"}, map[string]string{"locations": "enc:abc"})
	run(0, 0, &TimezoneParams{PathText: "enc:abc", Samples: 10}, map[string]string{"path": "enc:abc", "samples": "10"})
	run(0, 0, &TimezoneParams{Path: []Point{eiffel, liberty}, Samples: 3},
		map[string]string{"path": "48.8584,2.2945|40.6892,-74.0445", "samples": "3"})
	run(0, 0, &TimezoneParams{Bounds: &BBox{40, -75, 41, -73}, Rows: 3, Cols: 4},
		map[string]string{"bounds": "40,-75,41,-73", "rows": "3", "cols": "4"})
	if _, err := c.TimezoneBatch(bg, []Point{eiffel}, &TimezoneBatchParams{Timestamp: 1767225600, Lang: "de"}); err != nil {
		t.Fatal(err)
	}
	eqQuery(t, rec.last().query, map[string]string{"locations": "48.8584,2.2945", "timestamp": "1767225600", "language": "de"})
	if _, err := c.TimezoneBatch(bg, nil, &TimezoneBatchParams{LocationsText: "enc:_p~iF~ps|U", Language: "fr", Pretty: true}); err != nil {
		t.Fatal(err)
	}
	eqQuery(t, rec.last().query, map[string]string{"locations": "enc:_p~iF~ps|U", "language": "fr", "pretty": "1"})
}

func TestElevationParameters(t *testing.T) {
	rec := newRecorder(replay(t, "elevation"))
	s := &setup{}
	c := s.client(rec)
	run := func(p *ElevationParams, want map[string]string) {
		t.Helper()
		if _, err := c.Elevation(bg, 0, 0, p); err != nil {
			t.Fatal(err)
		}
		eqQuery(t, rec.last().query, want)
	}
	run(&ElevationParams{Locations: []Point{eiffel, liberty}, Pretty: true},
		map[string]string{"locations": "48.8584,2.2945|40.6892,-74.0445", "pretty": "1"})
	run(&ElevationParams{LocationsText: "enc:abc"}, map[string]string{"locations": "enc:abc"})
	run(&ElevationParams{Path: []Point{eiffel, liberty}, Samples: 5},
		map[string]string{"path": "48.8584,2.2945|40.6892,-74.0445", "samples": "5"})
	run(&ElevationParams{PathText: "enc:abc", Samples: 5}, map[string]string{"path": "enc:abc", "samples": "5"})
	run(&ElevationParams{Bounds: &BBox{27, 86, 28, 87}, Rows: 2, Cols: 2},
		map[string]string{"bounds": "27,86,28,87", "rows": "2", "cols": "2"})
	run(nil, map[string]string{"lat": "0", "lon": "0"})
}

func TestPostcodeParameters(t *testing.T) {
	rec := newRecorder(replay(t, "postcode"))
	s := &setup{}
	if _, err := s.client(rec).Postcode(bg, "75008", &PostcodeParams{Country: "FR", Pretty: true}); err != nil {
		t.Fatal(err)
	}
	eqQuery(t, rec.last().query, map[string]string{"code": "75008", "country": "FR", "pretty": "1"})
}

func TestBatchOptionsGoInTheQueryString(t *testing.T) {
	rec := newRecorder(replay(t, "forward_batch"))
	s := &setup{}
	c := s.client(rec)
	_, err := c.ForwardBatch(bg, []string{"Brandenburg Gate, Berlin"}, &ForwardBatchParams{
		Limit: 1, Countries: []string{"de"}, Bounds: &BBox{52, 13, 53, 14}, Strict: true, Proximity: &Point{52.5, 13.4},
		Types: []string{"poi"}, Lang: "fr", Annotations: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rec.last().json(t), map[string]interface{}{"queries": []interface{}{"Brandenburg Gate, Berlin"}}) {
		t.Fatalf("%s", rec.last().body)
	}
	eqQuery(t, rec.last().query, map[string]string{
		"limit": "1", "countries": "de", "bounds": "52,13,53,14", "strict": "1", "proximity": "52.5,13.4",
		"types": "poi", "lang": "fr", "annotations": "1",
	})
	if _, err := c.ReverseBatch(bg, []Point{eiffel}, &ReverseParams{Level: "street", Radius: 100, Boundaries: true}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rec.last().json(t), map[string]interface{}{"coords": []interface{}{[]interface{}{48.8584, 2.2945}}}) {
		t.Fatalf("%s", rec.last().body)
	}
	eqQuery(t, rec.last().query, map[string]string{"level": "street", "radius": "100", "lang": "en", "boundaries": "1"})
	if _, err := c.ElevationBatch(bg, []Point{eiffel}, &ElevationBatchParams{Pretty: true}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rec.last().json(t), map[string]interface{}{"locations": []interface{}{[]interface{}{48.8584, 2.2945}}}) {
		t.Fatalf("%s", rec.last().body)
	}
	eqQuery(t, rec.last().query, map[string]string{"pretty": "1"})
	if rec.last().header.Get("Content-Type") != "application/json" {
		t.Fatal("no JSON content type on a POST")
	}
}

func TestParamsPassThroughAndWin(t *testing.T) {
	rec := newRecorder(replay(t, "forward"))
	s := &setup{}
	_, err := s.client(rec).Forward(bg, "Brandenburg Gate, Berlin", &ForwardParams{Limit: 2, Params: Params{"limit": "5", "brand_new": "x", "lang": "it"}})
	if err != nil {
		t.Fatal(err)
	}
	q := rec.last().query
	if q["limit"] != "5" || q["brand_new"] != "x" || q["lang"] != "it" {
		t.Fatal(q)
	}
}

func TestEveryParamsStructHasPrettyAndParams(t *testing.T) {
	for _, p := range []interface{}{ForwardParams{}, ForwardBatchParams{}, ReverseParams{}, AutocompleteParams{}, IPParams{},
		TimezoneParams{}, TimezoneBatchParams{}, ElevationParams{}, ElevationBatchParams{}, PostcodeParams{}} {
		tp := reflect.TypeOf(p)
		if f, ok := tp.FieldByName("Pretty"); !ok || f.Tag.Get("q") != "pretty" {
			t.Fatalf("%s has no Pretty", tp.Name())
		}
		if f, ok := tp.FieldByName("Params"); !ok || f.Type != reflect.TypeOf(Params{}) {
			t.Fatalf("%s has no Params", tp.Name())
		}
	}
}
