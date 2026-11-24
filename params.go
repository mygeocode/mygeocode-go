package mygeocode

// Code below lists every documented parameter of every endpoint. The q tag is the API's name of the parameter;
// "-" marks the fields a method sends itself (lang, language, LocationsText of TimezoneBatch, Params).

// ForwardParams are the options of Forward. Besides the free text query, an address can be searched by its structured fields, or a result looked up again by its PlaceID.
type ForwardParams struct {
	// Street is the street with the house number, such as "10 Downing Street".
	Street string `q:"street"`
	// City is the city.
	City string `q:"city"`
	// State is the state, province or region.
	State string `q:"state"`
	// Postcode is the postal code.
	Postcode string `q:"postcode"`
	// Country is the country of the structured search, an ISO 3166-1 alpha-2 code.
	Country string `q:"country"`
	// PlaceID looks up the place_id of an earlier result.
	PlaceID string `q:"place_id"`
	// Limit is the number of results, 1 to 10 (zero: the API's default).
	Limit int `q:"limit"`
	// Countries limits the results to these ISO 3166-1 alpha-2 codes.
	Countries []string `q:"countries"`
	// Bounds ranks results inside the box first.
	Bounds *BBox `q:"bounds"`
	// Strict leaves out results outside Bounds.
	Strict bool `q:"strict"`
	// Proximity ranks results near the point first.
	Proximity *Point `q:"proximity"`
	// Types limits the results to these kinds: address, street, postcode, city, region, country, poi.
	Types []string `q:"types"`
	// Lang is the language of the results for this call (empty: the client's, English unless set).
	Lang string `q:"-"`
	// Annotations adds every extra block below.
	Annotations bool `q:"annotations"`
	// Timezone adds the timezone block.
	Timezone bool `q:"timezone"`
	// Elevation adds elevation_m.
	Elevation bool `q:"elevation"`
	// Wikidata adds the Wikidata id.
	Wikidata bool `q:"wikidata"`
	// CountryInfo adds the country block.
	CountryInfo bool `q:"country_info"`
	// Codes adds the codes block (plus code, geohash, MGRS ...).
	Codes bool `q:"codes"`
	// Sun adds sunrise and sunset.
	Sun bool `q:"sun"`
	// Regions adds the regions block.
	Regions bool `q:"regions"`
	// Districts adds the districts block.
	Districts bool `q:"districts"`
	// Census adds the census block.
	Census bool `q:"census"`
	// Demographics adds the demographics block.
	Demographics bool `q:"demographics"`
	// Road adds the road block.
	Road bool `q:"road"`
	// Site adds the site block.
	Site bool `q:"site"`
	// Pretty asks the API to indent its JSON.
	Pretty bool `q:"pretty"`
	// Params are extra query parameters, sent as they are; a name here replaces the value the library would send.
	Params Params `q:"-"`
}

// ForwardBatchParams are the options of ForwardBatch. They apply to every search in the batch.
type ForwardBatchParams struct {
	// Limit is the number of results, 1 to 10 (zero: the API's default).
	Limit int `q:"limit"`
	// Countries limits the results to these ISO 3166-1 alpha-2 codes.
	Countries []string `q:"countries"`
	// Bounds ranks results inside the box first.
	Bounds *BBox `q:"bounds"`
	// Strict leaves out results outside Bounds.
	Strict bool `q:"strict"`
	// Proximity ranks results near the point first.
	Proximity *Point `q:"proximity"`
	// Types limits the results to these kinds: address, street, postcode, city, region, country, poi.
	Types []string `q:"types"`
	// Lang is the language of the results for this call (empty: the client's, English unless set).
	Lang string `q:"-"`
	// Annotations adds every extra block below.
	Annotations bool `q:"annotations"`
	// Timezone adds the timezone block.
	Timezone bool `q:"timezone"`
	// Elevation adds elevation_m.
	Elevation bool `q:"elevation"`
	// Wikidata adds the Wikidata id.
	Wikidata bool `q:"wikidata"`
	// CountryInfo adds the country block.
	CountryInfo bool `q:"country_info"`
	// Codes adds the codes block (plus code, geohash, MGRS ...).
	Codes bool `q:"codes"`
	// Sun adds sunrise and sunset.
	Sun bool `q:"sun"`
	// Regions adds the regions block.
	Regions bool `q:"regions"`
	// Districts adds the districts block.
	Districts bool `q:"districts"`
	// Census adds the census block.
	Census bool `q:"census"`
	// Demographics adds the demographics block.
	Demographics bool `q:"demographics"`
	// Road adds the road block.
	Road bool `q:"road"`
	// Site adds the site block.
	Site bool `q:"site"`
	// Pretty asks the API to indent its JSON.
	Pretty bool `q:"pretty"`
	// Params are extra query parameters, sent as they are; a name here replaces the value the library would send.
	Params Params `q:"-"`
}

// ReverseParams are the options of Reverse and ReverseBatch.
type ReverseParams struct {
	// Level is how fine the answer is: address (default), street, postcode, city, region or country.
	Level string `q:"level"`
	// Radius is how far to look for an address, 1 to 5000 metres (zero: the API's default).
	Radius int `q:"radius"`
	// Lang is the language of the result for this call (empty: the client's, English unless set).
	Lang string `q:"-"`
	// Boundaries adds the administrative boundaries the point falls in (Australia).
	Boundaries bool `q:"boundaries"`
	// Annotations adds every extra block below.
	Annotations bool `q:"annotations"`
	// Timezone adds the timezone block.
	Timezone bool `q:"timezone"`
	// Elevation adds elevation_m.
	Elevation bool `q:"elevation"`
	// Wikidata adds the Wikidata id.
	Wikidata bool `q:"wikidata"`
	// CountryInfo adds the country block.
	CountryInfo bool `q:"country_info"`
	// Codes adds the codes block (plus code, geohash, MGRS ...).
	Codes bool `q:"codes"`
	// Sun adds sunrise and sunset.
	Sun bool `q:"sun"`
	// Regions adds the regions block.
	Regions bool `q:"regions"`
	// Districts adds the districts block.
	Districts bool `q:"districts"`
	// Census adds the census block.
	Census bool `q:"census"`
	// Demographics adds the demographics block.
	Demographics bool `q:"demographics"`
	// Road adds the road block.
	Road bool `q:"road"`
	// Site adds the site block.
	Site bool `q:"site"`
	// Pretty asks the API to indent its JSON.
	Pretty bool `q:"pretty"`
	// Params are extra query parameters, sent as they are; a name here replaces the value the library would send.
	Params Params `q:"-"`
}

// AutocompleteParams are the options of Autocomplete.
type AutocompleteParams struct {
	// Limit is the number of suggestions, 1 to 10 (zero: the API's default).
	Limit int `q:"limit"`
	// Countries limits the suggestions to these ISO 3166-1 alpha-2 codes.
	Countries []string `q:"countries"`
	// Bounds ranks suggestions inside the box first.
	Bounds *BBox `q:"bounds"`
	// Strict leaves out suggestions outside Bounds.
	Strict bool `q:"strict"`
	// Proximity ranks suggestions near the point first.
	Proximity *Point `q:"proximity"`
	// Types limits the suggestions to these kinds: address, street, postcode, city, region, country, poi.
	Types []string `q:"types"`
	// Lang is the language of the suggestions for this call (empty: the client's, English unless set).
	Lang string `q:"-"`
	// Pretty asks the API to indent its JSON.
	Pretty bool `q:"pretty"`
	// Params are extra query parameters, sent as they are; a name here replaces the value the library would send.
	Params Params `q:"-"`
}

// IPParams are the options of IP, IPv4, IPv6 and their batch forms.
type IPParams struct {
	// Elevation adds elevation_m for the location of each address.
	Elevation bool `q:"elevation"`
	// Pretty asks the API to indent its JSON.
	Pretty bool `q:"pretty"`
	// Params are extra query parameters, sent as they are; a name here replaces the value the library would send.
	Params Params `q:"-"`
}

// TimezoneParams are the options of Timezone. Tz, Locations, Path or a Bounds grid replace lat and lon, which are then not sent.
type TimezoneParams struct {
	// Tz asks by IANA zone id, such as "Europe/Paris", instead of by point.
	Tz string `q:"tz"`
	// Timestamp is the Unix time to answer for (zero: now).
	Timestamp int64 `q:"timestamp"`
	// Lang is the language of the zone names (empty: the client's lang if you set one, else English).
	Lang string `q:"-"`
	// Language is the API's own name for Lang; Lang wins when both are set.
	Language string `q:"-"`
	// Locations asks for several points at once.
	Locations []Point `q:"locations"`
	// LocationsText is Locations as the API's own text: "lat,lon|lat,lon" or an encoded polyline "enc:...".
	LocationsText string `q:"locations"`
	// Path is a line along which Samples points are taken.
	Path []Point `q:"path"`
	// PathText is Path as the API's own text: "lat,lon|lat,lon" or an encoded polyline "enc:...".
	PathText string `q:"path"`
	// Samples is how many points to take along Path, 2 to 100.
	Samples int `q:"samples"`
	// Bounds is a box covered by a grid of Rows by Cols points.
	Bounds *BBox `q:"bounds"`
	// Rows of the grid.
	Rows int `q:"rows"`
	// Cols of the grid.
	Cols int `q:"cols"`
	// Pretty asks the API to indent its JSON.
	Pretty bool `q:"pretty"`
	// Params are extra query parameters, sent as they are; a name here replaces the value the library would send.
	Params Params `q:"-"`
}

// TimezoneBatchParams are the options of TimezoneBatch.
type TimezoneBatchParams struct {
	// LocationsText sends the points as the API's own text ("lat,lon|lat,lon" or an encoded polyline "enc:...") instead of the points argument.
	LocationsText string `q:"-"`
	// Timestamp is the Unix time to answer for (zero: now).
	Timestamp int64 `q:"timestamp"`
	// Lang is the language of the zone names (empty: the client's lang if you set one, else English).
	Lang string `q:"-"`
	// Language is the API's own name for Lang; Lang wins when both are set.
	Language string `q:"-"`
	// Pretty asks the API to indent its JSON.
	Pretty bool `q:"pretty"`
	// Params are extra query parameters, sent as they are; a name here replaces the value the library would send.
	Params Params `q:"-"`
}

// ElevationParams are the options of Elevation. Locations, Path or a Bounds grid replace lat and lon, which are then not sent.
type ElevationParams struct {
	// Locations asks for several points at once.
	Locations []Point `q:"locations"`
	// LocationsText is Locations as the API's own text: "lat,lon|lat,lon" or an encoded polyline "enc:...".
	LocationsText string `q:"locations"`
	// Path is a line along which Samples points are taken.
	Path []Point `q:"path"`
	// PathText is Path as the API's own text: "lat,lon|lat,lon" or an encoded polyline "enc:...".
	PathText string `q:"path"`
	// Samples is how many points to take along Path, 2 to 100.
	Samples int `q:"samples"`
	// Bounds is a box covered by a grid of Rows by Cols points.
	Bounds *BBox `q:"bounds"`
	// Rows of the grid.
	Rows int `q:"rows"`
	// Cols of the grid.
	Cols int `q:"cols"`
	// Pretty asks the API to indent its JSON.
	Pretty bool `q:"pretty"`
	// Params are extra query parameters, sent as they are; a name here replaces the value the library would send.
	Params Params `q:"-"`
}

// ElevationBatchParams are the options of ElevationBatch.
type ElevationBatchParams struct {
	// Pretty asks the API to indent its JSON.
	Pretty bool `q:"pretty"`
	// Params are extra query parameters, sent as they are; a name here replaces the value the library would send.
	Params Params `q:"-"`
}

// PostcodeParams are the options of Postcode.
type PostcodeParams struct {
	// Country narrows the lookup to one country, an ISO 3166-1 alpha-2 code.
	Country string `q:"country"`
	// Pretty asks the API to indent its JSON.
	Pretty bool `q:"pretty"`
	// Params are extra query parameters, sent as they are; a name here replaces the value the library would send.
	Params Params `q:"-"`
}
