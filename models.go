package mygeocode

import "encoding/json"

// The result types. Every field can be missing, so text, numbers and flags are pointers (nil when the API did not
// send the field or sent null) and lists are slices (empty when missing). A field this library reads is never
// renamed or removed by the API; at most it becomes nil. Raw holds the JSON object exactly as the API sent it, fields
// this version does not know included. StringValue, FloatValue, IntValue and BoolValue read a pointer with a default.

// Bounds is the box around a place, in degrees.
type Bounds struct {
	// Northern edge.
	North *float64 `json:"north"`
	// Southern edge.
	South *float64 `json:"south"`
	// Eastern edge.
	East *float64 `json:"east"`
	// Western edge.
	West *float64 `json:"west"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// Components is the parts of an address. Each is text (a number the API sends in one of these fields is read as text), nil when the address has no such part.
type Components struct {
	// Point of interest name.
	POI *string `json:"poi"`
	// Building name.
	Building *string `json:"building"`
	// Flat, suite or unit.
	Unit *string `json:"unit"`
	// Floor.
	Level *string `json:"level"`
	// House number.
	HouseNumber *string `json:"house_number"`
	// Street.
	Road *string `json:"road"`
	// Lot or plot.
	Lot *string `json:"lot"`
	// Suburb or neighbourhood.
	Suburb *string `json:"suburb"`
	// Municipality.
	Municipality *string `json:"municipality"`
	// City, town or village.
	City *string `json:"city"`
	// County or district.
	County *string `json:"county"`
	// State, province or region.
	State *string `json:"state"`
	// State code.
	StateCode *string `json:"state_code"`
	// Postal code.
	Postcode *string `json:"postcode"`
	// Country name.
	Country *string `json:"country"`
	// ISO 3166-1 alpha-2 country code, lower case.
	CountryCode *string `json:"country_code"`
	// Official name of the place.
	OfficialName *string `json:"official_name"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// Place is one forward or reverse result. The extra blocks you ask for with the annotation flags are raw JSON objects, nil when not asked for.
type Place struct {
	// The whole address on one line.
	Formatted *string `json:"formatted"`
	// Latitude.
	Lat *float64 `json:"lat"`
	// Longitude.
	Lon *float64 `json:"lon"`
	// The kind of result: address, street, postcode, city, suburb, region, country or poi.
	Type *string `json:"type"`
	// How precise the match is: house, street, postcode, city, region or country.
	Precision *string `json:"precision"`
	// How sure the match is, 0 to 1.
	Confidence *float64 `json:"confidence"`
	// An id you can look up again with ForwardParams.PlaceID.
	PlaceID *string `json:"place_id"`
	// The address parts.
	Components *Components `json:"components"`
	// The box around the place.
	Bounds *Bounds `json:"bounds"`
	// Reverse only: metres from the point you gave.
	DistanceM *int64 `json:"distance_m"`
	// The place name on one line, without street and number.
	FormattedPlace *string `json:"formatted_place"`
	// The place block.
	Place json.RawMessage `json:"place"`
	// The land parcel block, where the country publishes one.
	Parcel json.RawMessage `json:"parcel"`
	// The codes block (plus code, geohash, MGRS ...), with the Codes flag.
	Codes json.RawMessage `json:"codes"`
	// The population of a city or region result.
	Population *int64 `json:"population"`
	// The UN/LOCODE block.
	UNLocode json.RawMessage `json:"un_locode"`
	// The timezone block, with the Timezone flag.
	Timezone json.RawMessage `json:"timezone"`
	// Metres above sea level, with the Elevation flag.
	ElevationM *float64 `json:"elevation_m"`
	// The Wikidata id, with the Wikidata flag.
	Wikidata *string `json:"wikidata"`
	// The country block, with the CountryInfo flag.
	CountryInfo json.RawMessage `json:"country_info"`
	// Sunrise and sunset, with the Sun flag.
	Sun json.RawMessage `json:"sun"`
	// The regions block, with the Regions flag.
	Regions json.RawMessage `json:"regions"`
	// The districts block, with the Districts flag.
	Districts json.RawMessage `json:"districts"`
	// The census block, with the Census flag.
	Census json.RawMessage `json:"census"`
	// The demographics block, with the Demographics flag.
	Demographics json.RawMessage `json:"demographics"`
	// The road block, with the Road flag.
	Road json.RawMessage `json:"road"`
	// The nearest intersection.
	Intersection json.RawMessage `json:"intersection"`
	// The site block, with the Site flag.
	Site json.RawMessage `json:"site"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// ForwardResponse is the answer of Forward: the matches, best first.
type ForwardResponse struct {
	// "ok".
	Status *string `json:"status"`
	// The text that was searched.
	Query *string `json:"query"`
	// The matches, best first; empty when nothing matched.
	Results []Place `json:"results"`
	// Meta holds the request id and quota figures from the response headers (empty on an entry inside a batch).
	Meta Meta `json:"-"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// ForwardBatchEntry is one search of a forward batch.
type ForwardBatchEntry struct {
	// The text that was searched.
	Query *string `json:"query"`
	// Its matches, best first.
	Results []Place `json:"results"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// ForwardBatchResponse is the answer of ForwardBatch.
type ForwardBatchResponse struct {
	// "ok".
	Status *string `json:"status"`
	// One entry per search, in the order sent.
	Results []ForwardBatchEntry `json:"results"`
	// Meta holds the request id and quota figures from the response headers (empty on an entry inside a batch).
	Meta Meta `json:"-"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// ReverseResponse is the answer of Reverse, and one entry of a reverse batch.
type ReverseResponse struct {
	// "ok".
	Status *string `json:"status"`
	// The latitude you sent.
	Lat *float64 `json:"lat"`
	// The longitude you sent.
	Lon *float64 `json:"lon"`
	// The nearest address, or nil when nothing is in range.
	Result *Place `json:"result"`
	// The Australian boundaries, with ReverseParams.Boundaries.
	Boundaries json.RawMessage `json:"boundaries"`
	// Meta holds the request id and quota figures from the response headers (empty on an entry inside a batch).
	Meta Meta `json:"-"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// ReverseBatchResponse is the answer of ReverseBatch.
type ReverseBatchResponse struct {
	// "ok".
	Status *string `json:"status"`
	// One entry per point, in the order sent.
	Results []ReverseResponse `json:"results"`
	// Meta holds the request id and quota figures from the response headers (empty on an entry inside a batch).
	Meta Meta `json:"-"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// Suggestion is one autocomplete suggestion.
type Suggestion struct {
	// The suggestion as shown to the user.
	Text *string `json:"text"`
	// Latitude.
	Lat *float64 `json:"lat"`
	// Longitude.
	Lon *float64 `json:"lon"`
	// The kind of result.
	Type *string `json:"type"`
	// How precise the match is.
	Precision *string `json:"precision"`
	// Look it up with ForwardParams.PlaceID for the full result.
	PlaceID *string `json:"place_id"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// AutocompleteResponse is the answer of Autocomplete.
type AutocompleteResponse struct {
	// "ok".
	Status *string `json:"status"`
	// The text that was looked up.
	Query *string `json:"query"`
	// The suggestions, best first.
	Suggestions []Suggestion `json:"suggestions"`
	// Meta holds the request id and quota figures from the response headers (empty on an entry inside a batch).
	Meta Meta `json:"-"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// ErrorDetail says why the API rejected one entry of a batch.
type ErrorDetail struct {
	// The error code, such as invalid_ip.
	Code *string `json:"code"`
	// The explanation.
	Message *string `json:"message"`
	// The parameter at fault.
	Param *string `json:"param"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// IPInfo is an IPv4 or IPv6 lookup, and one entry of an IP batch. In a batch, an address the API cannot read has the status "error" and an Error.
type IPInfo struct {
	// "ok", or "error" on a batch entry the API could not read.
	Status *string `json:"status"`
	// The address.
	IP *string `json:"ip"`
	// The IP version, 4 or 6.
	Version *int64 `json:"version"`
	// Whether the address is in the data.
	Found *bool `json:"found"`
	// The network the address belongs to.
	Prefix *string `json:"prefix"`
	// Country name.
	Country *string `json:"country"`
	// ISO 3166-1 alpha-2 country code.
	CountryCode *string `json:"country_code"`
	// Continent name.
	Continent *string `json:"continent"`
	// Continent code.
	ContinentCode *string `json:"continent_code"`
	// Region name.
	Region *string `json:"region"`
	// Region code.
	RegionCode *string `json:"region_code"`
	// City.
	City *string `json:"city"`
	// Postal code.
	Postcode *string `json:"postcode"`
	// Latitude.
	Lat *float64 `json:"lat"`
	// Longitude.
	Lon *float64 `json:"lon"`
	// Metres above sea level, with IPParams.Elevation.
	ElevationM *float64 `json:"elevation_m"`
	// IANA timezone id.
	Timezone *string `json:"timezone"`
	// UTC offset, such as -07:00.
	UTCOffset *string `json:"utc_offset"`
	// UTC offset in seconds.
	UTCOffsetSeconds *int64 `json:"utc_offset_seconds"`
	// Whether the country is in the European Union.
	IsEU *bool `json:"is_eu"`
	// Autonomous system number of the network owner.
	ASN *int64 `json:"asn"`
	// Name of the network owner.
	ASName *string `json:"as_name"`
	// Organisation.
	Org *string `json:"org"`
	// What the network is used for.
	UsageType *string `json:"usage_type"`
	// The regional internet registry.
	Registry *string `json:"registry"`
	// When the block was allocated.
	Allocated *string `json:"allocated"`
	// Tags such as anycast or datacenter.
	Tags []string `json:"tags"`
	// Anycast network.
	IsAnycast *bool `json:"is_anycast"`
	// Satellite network.
	IsSatellite *bool `json:"is_satellite"`
	// Data centre network.
	IsDatacenter *bool `json:"is_datacenter"`
	// Known proxy.
	IsProxy *bool `json:"is_proxy"`
	// Network of a VPN provider.
	IsVPNProvider *bool `json:"is_vpn_provider"`
	// Tor node.
	IsTor *bool `json:"is_tor"`
	// The crawler block, when the address belongs to a known crawler.
	Crawler json.RawMessage `json:"crawler"`
	// Private or reserved address.
	IsPrivate *bool `json:"is_private"`
	// Why a private or reserved address has no data.
	Reason *string `json:"reason"`
	// Abuse contact of the network.
	AbuseEmail *string `json:"abuse_email"`
	// The threat block.
	Threat json.RawMessage `json:"threat"`
	// When the record was last updated.
	UpdatedAt *string `json:"updated_at"`
	// On a batch entry the API rejected: why.
	Error *ErrorDetail `json:"error"`
	// Meta holds the request id and quota figures from the response headers (empty on an entry inside a batch).
	Meta Meta `json:"-"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// IPBatchResponse is the answer of the IP batch methods.
type IPBatchResponse struct {
	// "ok".
	Status *string `json:"status"`
	// One entry per address, in the order sent.
	Results []IPInfo `json:"results"`
	// Meta holds the request id and quota figures from the response headers (empty on an entry inside a batch).
	Meta Meta `json:"-"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// TimezoneTransition is the next change of UTC offset, such as the start or end of daylight saving time.
type TimezoneTransition struct {
	// Unix time of the change.
	Timestamp *int64 `json:"timestamp"`
	// The offset after the change.
	UTCOffset *string `json:"utc_offset"`
	// The offset after the change, in seconds.
	UTCOffsetSeconds *int64 `json:"utc_offset_seconds"`
	// Whether daylight saving time applies after the change.
	DST *bool `json:"dst"`
	// The abbreviation after the change.
	Abbreviation *string `json:"abbreviation"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// TimezoneInfo is the answer of Timezone, and one entry of a timezone batch.
type TimezoneInfo struct {
	// "ok".
	Status *string `json:"status"`
	// Latitude; nil when you asked by zone id.
	Lat *float64 `json:"lat"`
	// Longitude; nil when you asked by zone id.
	Lon *float64 `json:"lon"`
	// IANA timezone id, such as Asia/Tokyo.
	Timezone *string `json:"timezone"`
	// The zone name, such as Japan Standard Time.
	Name *string `json:"name"`
	// The abbreviation, such as JST.
	Abbreviation *string `json:"abbreviation"`
	// The UTC offset, such as +09:00.
	UTCOffset *string `json:"utc_offset"`
	// The offset in seconds.
	UTCOffsetSeconds *int64 `json:"utc_offset_seconds"`
	// The offset without daylight saving time.
	RawOffsetSeconds *int64 `json:"raw_offset_seconds"`
	// Whether daylight saving time applies.
	DST *bool `json:"dst"`
	// The daylight saving part of the offset.
	DSTOffsetSeconds *int64 `json:"dst_offset_seconds"`
	// The Unix time answered for.
	Timestamp *int64 `json:"timestamp"`
	// Local time, ISO 8601.
	LocalTime *string `json:"local_time"`
	// ISO 3166-1 alpha-2 country code.
	CountryCode *string `json:"country_code"`
	// Whether the point is at sea.
	IsOcean *bool `json:"is_ocean"`
	// The Windows zone id.
	WindowsTimezone *string `json:"windows_timezone"`
	// The next change of offset.
	NextTransition *TimezoneTransition `json:"next_transition"`
	// Meta holds the request id and quota figures from the response headers (empty on an entry inside a batch).
	Meta Meta `json:"-"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// TimezoneBatchResponse is the answer of TimezoneBatch.
type TimezoneBatchResponse struct {
	// "ok".
	Status *string `json:"status"`
	// The Unix time answered for.
	Timestamp *int64 `json:"timestamp"`
	// One entry per point, in the order sent.
	Results []TimezoneInfo `json:"results"`
	// Meta holds the request id and quota figures from the response headers (empty on an entry inside a batch).
	Meta Meta `json:"-"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// ElevationPoint is the height of one point.
type ElevationPoint struct {
	// Latitude.
	Lat *float64 `json:"lat"`
	// Longitude.
	Lon *float64 `json:"lon"`
	// Metres above sea level (negative for the sea floor), nil where there is no data.
	ElevationM *float64 `json:"elevation_m"`
	// The size of the grid cell, in metres.
	ResolutionM *float64 `json:"resolution_m"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// ElevationResponse is the answer of the elevation methods. It always has Results, one entry for a single point.
type ElevationResponse struct {
	// "ok".
	Status *string `json:"status"`
	// One entry per point, in order.
	Results []ElevationPoint `json:"results"`
	// Meta holds the request id and quota figures from the response headers (empty on an entry inside a batch).
	Meta Meta `json:"-"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// PostcodePlace is one place that uses a postal code.
type PostcodePlace struct {
	// The postal code, as text.
	Postcode *string `json:"postcode"`
	// The place name.
	Place *string `json:"place"`
	// Region name.
	Region *string `json:"region"`
	// Region code.
	RegionCode *string `json:"region_code"`
	// Country name.
	Country *string `json:"country"`
	// ISO 3166-1 alpha-2 country code.
	CountryCode *string `json:"country_code"`
	// Latitude.
	Lat *float64 `json:"lat"`
	// Longitude.
	Lon *float64 `json:"lon"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}

// PostcodeResponse is the answer of Postcode.
type PostcodeResponse struct {
	// "ok".
	Status *string `json:"status"`
	// The code that was looked up, as text.
	Postcode *string `json:"postcode"`
	// The country it was matched in.
	CountryCode *string `json:"country_code"`
	// Every place that uses the code.
	Results []PostcodePlace `json:"results"`
	// Meta holds the request id and quota figures from the response headers (empty on an entry inside a batch).
	Meta Meta `json:"-"`
	// Raw is the JSON object exactly as the API sent it.
	Raw json.RawMessage `json:"-"`
}
