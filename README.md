# My Geocode Go SDK

The official Go client for the [My Geocode API](https://www.mygeocode.com/): forward and reverse geocoding, address
autocomplete, IPv4 and IPv6 lookup, timezones, elevation and postal codes.

![Go 1.13 and newer](https://img.shields.io/badge/go-1.13%20and%20newer-00add8)
![Dependencies: none](https://img.shields.io/badge/dependencies-none-brightgreen)
![License: MIT](https://img.shields.io/badge/license-MIT-blue)

- Standard library only: `net/http` and `encoding/json`, nothing to vendor.
- Works on Go 1.13 and newer, with a `context.Context` on every call. Safe to share between goroutines.
- Nothing global: no logger, no environment changes, no `init` work. The default transport is a fresh `http.Client`.
- 2,500 free requests a day without a key or an account.

## Install

```bash
go get github.com/mygeocode/mygeocode-go
```

```go
import "github.com/mygeocode/mygeocode-go" // package mygeocode
```

## Quick start (no key)

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mygeocode/mygeocode-go"
)

func main() {
	ctx := context.Background()
	c := mygeocode.NewClient()

	found, err := c.Forward(ctx, "Brandenburg Gate, Berlin", nil)
	if err != nil {
		log.Fatal(err)
	}
	place := found.Results[0]
	fmt.Println(mygeocode.StringValue(place.Formatted)) // Brandenburg Gate, State of Berlin, Germany
	fmt.Println(*place.Lat, *place.Lon)                 // 52.516268 13.377692

	address, err := c.Reverse(ctx, 48.8584, 2.2945, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mygeocode.StringValue(address.Result.Formatted)) // 5 Avenue Anatole France, 75007 Paris, France
}
```

Without a key, each network gets 2,500 free requests a day. `found.Meta.QuotaFreeRemaining` tells you how many are
left. Make one client and share it: it is safe to use from many goroutines.

## Using a key

A key adds your account's prepaid credit or Unlimited package on top of the free requests. Create one in the
[dashboard](https://www.mygeocode.com/dashboard).

```go
c := mygeocode.NewClient(mygeocode.WithKey("YOUR_KEY"))
```

If you leave out `WithKey`, the client reads the `MYGEOCODE_KEY` environment variable. Pass `WithKey("")` to make
sure no key is sent. The key always travels in the `X-API-Key` header, never in the URL, and the SDK never writes it
anywhere.

## Language

Results are in English unless you set `lang`. Set it once on the client, or per call:

```go
c := mygeocode.NewClient(mygeocode.WithLang("de"))

c.Reverse(ctx, 48.8584, 2.2945, &mygeocode.ReverseParams{Lang: "fr"}) // this call only
```

`lang` takes an ISO 639-1 code such as `de`, `ja` or `pt-BR`. The whole address comes back in that language and
script. On its own, the API answers each result in the language of its area; the SDK sends `lang=en` by default so
that every result reads the same.

## Methods

Every method takes a `context.Context` first, then its main arguments, then a pointer to a params struct named after
the method, such as `*mygeocode.ForwardParams`. Pass `nil` for no options. The fields carry the API's own parameter
names, and each method has its own params type, so an option that does not belong to a method does not compile:

```go
c.Forward(ctx, "Brandenburg Gate, Berlin", &mygeocode.ForwardParams{Limit: 3, Countries: []string{"de", "at"}, Timezone: true})
```

A zero value means "not set": an empty string, a `0` or a `false` is not sent. Every params struct also has:

- `Params`: more query parameters (`mygeocode.Params{"name": "value"}`), sent as they are and winning over the named
  fields, so a parameter the API adds later works before the SDK knows it.
- `Pretty`: `true` for indented JSON.

Points are `mygeocode.Point{Lat: 48.8584, Lon: 2.2945}`, boxes `&mygeocode.BBox{South: ..., West: ..., North: ...,
East: ...}`.

| Method | Endpoint |
|---|---|
| [`Forward`](#forward) | `GET /v1/forward` |
| [`Reverse`](#reverse) | `GET /v1/reverse` |
| [`Autocomplete`](#autocomplete) | `GET /v1/autocomplete` |
| [`IP`, `IPv4`, `IPv6`](#ip-ipv4-ipv6) | `GET /v1/ip`, `/v1/ipv4`, `/v1/ipv6` |
| [`Timezone`](#timezone) | `GET /v1/timezone` |
| [`Elevation`](#elevation) | `GET /v1/elevation` |
| [`Postcode`](#postcode) | `GET /v1/postcode` |
| [`ForwardBatch`](#forwardbatch) | `POST /v1/forward` |
| [`ReverseBatch`](#reversebatch) | `POST /v1/reverse` |
| [`IPBatch`, `IPv4Batch`, `IPv6Batch`](#ipbatch-ipv4batch-ipv6batch) | `GET /v1/ip`, `/v1/ipv4`, `/v1/ipv6` with `ips` |
| [`TimezoneBatch`](#timezonebatch) | `GET /v1/timezone` with `locations` |
| [`ElevationBatch`](#elevationbatch) | `POST /v1/elevation` |

In the examples below, `c` is a client and `ctx` a context; checking `err` is left out to keep them short.

### Forward

Text, structured address fields or a place id to coordinates. Returns a `*ForwardResponse` whose `Results` holds up
to `Limit` `Place` values, best first. An empty list means nothing matched.

```go
found, err := c.Forward(ctx, "Brandenburg Gate, Berlin", nil)
fmt.Println(mygeocode.StringValue(found.Results[0].Formatted)) // Brandenburg Gate, State of Berlin, Germany

c.Forward(ctx, "Brandenburg Gate, Berlin", &mygeocode.ForwardParams{Limit: 3, Countries: []string{"de"}})
c.Forward(ctx, "", &mygeocode.ForwardParams{Street: "10 Downing Street", City: "London", Country: "GB"})
```

`ForwardParams`:

| Field | What it does |
|---|---|
| `Street`, `City`, `State`, `Postcode`, `Country` | Structured search instead of (or as well as) free text. Pass `""` as the query to search by these alone. |
| `PlaceID` | Look up a `place_id` from an earlier result. |
| `Limit` | 1 to 10 results. |
| `Countries` | ISO 3166-1 alpha-2 codes to search in, such as `[]string{"de", "at"}`. |
| `Bounds` | A box that results should fall in: `&mygeocode.BBox{South, West, North, East}`. |
| `Strict` | Only return results inside `Bounds`. |
| `Proximity` | Prefer results near this point: `&mygeocode.Point{Lat, Lon}`. |
| `Types` | Kinds of result to keep, such as `address` or `poi`. |
| `Lang` | Language of the result. |
| `Timezone`, `Elevation`, `Wikidata`, `CountryInfo`, `Codes`, `Sun`, `Regions`, `Districts`, `Census`, `Demographics`, `Road`, `Site` | Each adds its extra block to every result. |
| `Annotations` | Adds every extra block at once. |

### Reverse

A point to the nearest address. Returns a `*ReverseResponse` whose `Result` is a `*Place`, or `nil` when nothing is
in range.

```go
address, err := c.Reverse(ctx, 48.8584, 2.2945, nil)
place := address.Result
fmt.Println(mygeocode.StringValue(place.Formatted)) // 5 Avenue Anatole France, 75007 Paris, France
fmt.Println(mygeocode.IntValue(place.DistanceM))    // 47 (metres from the point you gave)

c.Reverse(ctx, 48.8584, 2.2945, &mygeocode.ReverseParams{Level: "city", Timezone: true})
```

`ReverseParams`:

| Field | What it does |
|---|---|
| `Level` | `address`, `street`, `postcode`, `city`, `region` or `country`. |
| `Radius` | Search radius in metres, 1 to 5000. |
| `Lang` | Language of the result. |
| `Boundaries` | Add the administrative boundaries the point falls in (Australia). |
| `Timezone`, `Elevation`, `Wikidata`, `CountryInfo`, `Codes`, `Sun`, `Regions`, `Districts`, `Census`, `Demographics`, `Road`, `Site` | Each adds its extra block, as for `Forward`. |
| `Annotations` | Adds every extra block at once. |

### Autocomplete

Suggestions while someone types an address. Returns an `*AutocompleteResponse` with a `Suggestions` list; each
`Suggestion` has `Text`, `Lat`, `Lon`, `Type`, `Precision` and `PlaceID`.

```go
typed, err := c.Autocomplete(ctx, "1600 Pennsylvania", nil)
fmt.Println(mygeocode.StringValue(typed.Suggestions[0].Text)) // 1600 Pennsylvania Avenue NW, Washington, DC 20500, United States

c.Autocomplete(ctx, "1600 Pennsylvania", &mygeocode.AutocompleteParams{Limit: 5, Countries: []string{"us"}})
```

`AutocompleteParams`: `Limit`, `Countries`, `Proximity`, `Bounds`, `Strict`, `Types` and `Lang`, as for `Forward`.

### IP, IPv4, IPv6

Country, region, city, coordinates, timezone, network owner (ASN) and proxy, VPN and data centre flags for an IP
address. `IP` takes either version; `IPv4` and `IPv6` accept only their own. An empty address looks up the address
the request comes from. Returns an `*IPInfo`.

```go
google, err := c.IP(ctx, "8.8.8.8", nil)
fmt.Println(mygeocode.StringValue(google.Country), mygeocode.StringValue(google.City)) // United States Mountain View
fmt.Println(mygeocode.IntValue(google.ASN), mygeocode.StringValue(google.ASName))     // 15169 Google LLC
fmt.Println(mygeocode.BoolValue(google.IsDatacenter))                                 // true

c.IPv6(ctx, "2001:4860:4860::8888", nil)
c.IP(ctx, "", nil)                                               // the address the request came from
c.IPv4(ctx, "8.8.8.8", &mygeocode.IPParams{Elevation: true}) // adds ElevationM
```

`IPParams`: `Elevation`. The full list of `IPInfo` fields is under [Results and meta](#results-and-meta).

### Timezone

The IANA zone, UTC offset, daylight saving state and local time for a point, now or at a given moment. Returns a
`*TimezoneInfo`.

```go
tokyo, err := c.Timezone(ctx, 35.6762, 139.6503, nil)
fmt.Println(mygeocode.StringValue(tokyo.Timezone), mygeocode.StringValue(tokyo.UTCOffset))  // Asia/Tokyo +09:00
fmt.Println(mygeocode.StringValue(tokyo.Name), mygeocode.StringValue(tokyo.Abbreviation))  // Japan Standard Time JST
fmt.Println(mygeocode.StringValue(tokyo.LocalTime))                                         // 2026-11-24T18:07:15+09:00

c.Timezone(ctx, 35.6762, 139.6503, &mygeocode.TimezoneParams{Timestamp: 1767225600})
c.Timezone(ctx, 0, 0, &mygeocode.TimezoneParams{Tz: "Europe/Paris"})
```

`TimezoneParams`:

| Field | What it does |
|---|---|
| `Tz` | Look up a zone by its IANA or Windows id instead of a point. `lat` and `lon` are then not sent (pass `0, 0`). |
| `Timestamp` | Unix time to answer for. Default: now. |
| `Lang` | Language of the zone name. Sent as the API's `language` parameter, only when you set it here or on the client. `Language` works as another name for this field. |
| `Locations`, `LocationsText` | Several points: a `[]mygeocode.Point`, or `"lat,lon\|lat,lon"` or an `enc:` polyline. |
| `Path`, `PathText`, `Samples` | Points along a path, sampled `Samples` times (2 to 100). |
| `Bounds`, `Rows`, `Cols` | A grid of points over a box. |

With `Tz`, `Locations`, `Path` or `Bounds` set, `lat` and `lon` are not sent. For a plain list of points,
`TimezoneBatch` is simpler.

### Elevation

Height in metres above sea level (negative for the sea floor), from a global 15 arc-second grid. Returns an
`*ElevationResponse` whose `Results` holds `ElevationPoint` values with `Lat`, `Lon`, `ElevationM` and `ResolutionM`.

```go
everest, err := c.Elevation(ctx, 27.9881, 86.9250, nil)
fmt.Println(mygeocode.FloatValue(everest.Results[0].ElevationM)) // 8627

path := []mygeocode.Point{{Lat: 27.9881, Lon: 86.9250}, {Lat: 28.0, Lon: 86.85}}
c.Elevation(ctx, 0, 0, &mygeocode.ElevationParams{Path: path, Samples: 10})
```

`ElevationParams`: `Locations`, `Path` and `Samples`, or `Bounds`, `Rows` and `Cols`, as for `Timezone`. With one of
them set, `lat` and `lon` are not sent.

### Postcode

A postal code to its place, region and coordinates. Returns a `*PostcodeResponse` whose `Results` lists every place
that uses the code.

```go
nyc, err := c.Postcode(ctx, "10001", &mygeocode.PostcodeParams{Country: "US"})
first := nyc.Results[0]
fmt.Println(mygeocode.StringValue(first.Place), mygeocode.StringValue(first.RegionCode)) // New York NY
fmt.Println(*first.Lat, *first.Lon)                                                        // 40.748401 -73.996702

c.Postcode(ctx, "75008", &mygeocode.PostcodeParams{Country: "FR"})
```

`PostcodeParams`: `Country` (ISO alpha-2). Without it, the code is matched in every country.

### ForwardBatch

Up to 100 searches in one request (each counts as one request against your quota). Returns a
`*ForwardBatchResponse` whose `Results` holds one `ForwardBatchEntry` per search, in the order you sent them; each
entry has the `Query` and its own `Results` list of `Place` values.

```go
batch, err := c.ForwardBatch(ctx, []string{"Brandenburg Gate, Berlin", "Sydney Opera House", "Statue of Liberty, New York"}, nil)
for _, entry := range batch.Results {
	fmt.Printf("%s: %s\n", mygeocode.StringValue(entry.Query), mygeocode.StringValue(entry.Results[0].Formatted))
}
// Brandenburg Gate, Berlin: Brandenburg Gate, State of Berlin, Germany
// Sydney Opera House: Sydney Opera House, Sydney, New South Wales, Australia
// Statue of Liberty, New York: Statue of Liberty, New York, New York, United States

c.ForwardBatch(ctx, []string{"Brandenburg Gate, Berlin", "Sydney Opera House"}, &mygeocode.ForwardBatchParams{Limit: 1, Timezone: true})
```

`ForwardBatchParams` apply to every search in the batch: `Limit`, `Countries`, `Bounds`, `Strict`, `Proximity`,
`Types`, `Lang` and the extra blocks, as for `Forward`. Structured fields and `PlaceID` are not available in a batch;
use `Forward` for those.

### ReverseBatch

Up to 100 points in one request (each counts as one request). Returns a `*ReverseBatchResponse` whose `Results` holds
one `ReverseResponse` per point, in the order you sent them; its `Result` is a `*Place`, or `nil` when nothing is in
range.

```go
points := []mygeocode.Point{{Lat: 48.8584, Lon: 2.2945}, {Lat: 40.6892, Lon: -74.0445}, {Lat: -33.8568, Lon: 151.2153}}
batch, err := c.ReverseBatch(ctx, points, nil)
for _, point := range batch.Results {
	if point.Result == nil {
		fmt.Println("nothing in range")
		continue
	}
	fmt.Println(mygeocode.StringValue(point.Result.Formatted))
}
// 5 Avenue Anatole France, 75007 Paris, France
// 1 Liberty Island, New York, NY 10004, United States
// 2B Macquarie Street, Sydney NSW 2000, Australia
```

It takes the same `ReverseParams` as `Reverse`, applied to every point in the batch.

### IPBatch, IPv4Batch, IPv6Batch

Up to 100 addresses in one request. `IPBatch` takes both versions mixed; `IPv4Batch` and `IPv6Batch` accept only
their own. Returns an `*IPBatchResponse` whose `Results` holds one `IPInfo` per address, in the order you sent them.
An address the API cannot read gets an entry whose `Status` is `error` and whose `Error` explains why, while the rest
still answer.

```go
batch, err := c.IPv4Batch(ctx, []string{"8.8.8.8", "1.1.1.1", "9.9.9.9"}, nil)
for _, info := range batch.Results {
	fmt.Printf("%s: %s\n", mygeocode.StringValue(info.IP), mygeocode.StringValue(info.ASName))
}
// 8.8.8.8: Google LLC
// 1.1.1.1: Cloudflare, Inc.
// 9.9.9.9: Quad9

c.IPv6Batch(ctx, []string{"2001:4860:4860::8888", "2606:4700:4700::1111"}, nil)
c.IPBatch(ctx, []string{"8.8.8.8", "2606:4700:4700::1111"}, &mygeocode.IPParams{Elevation: true})
```

`IPParams`: `Elevation`, adds `ElevationM` to every address.

### TimezoneBatch

Timezones for up to 100 points. Returns a `*TimezoneBatchResponse` whose `Results` holds one `TimezoneInfo` per
point, in the order you sent them.

```go
points := []mygeocode.Point{{Lat: 48.8584, Lon: 2.2945}, {Lat: 40.6892, Lon: -74.0445}, {Lat: -33.8568, Lon: 151.2153}}
batch, err := c.TimezoneBatch(ctx, points, nil)
for _, zone := range batch.Results {
	fmt.Println(mygeocode.StringValue(zone.Timezone))
}
// Europe/Paris
// America/New_York
// Australia/Sydney
```

`TimezoneBatchParams`: `Timestamp` (Unix time, default now) and `Lang` (also accepted as `Language`), as for
`Timezone`, and `LocationsText` to send the points as an `enc:` polyline instead (the points argument may then be
`nil`).

### ElevationBatch

Heights for up to 100 points, sent as a POST body. Returns an `*ElevationResponse`.

```go
points := []mygeocode.Point{{Lat: 48.8584, Lon: 2.2945}, {Lat: 40.6892, Lon: -74.0445}, {Lat: -33.8568, Lon: 151.2153}}
batch, err := c.ElevationBatch(ctx, points, nil)
for _, point := range batch.Results {
	fmt.Println(mygeocode.FloatValue(point.ElevationM))
}
// 37
// 0
// 10
```

## Results and meta

Results are plain structs whose JSON tags carry the API's own field names, so the
[API documentation](https://www.mygeocode.com/docs/) describes them field by field (`place_id` is `PlaceID`,
`as_name` is `ASName`):

```go
found, err := c.Forward(ctx, "Brandenburg Gate, Berlin", nil)
place := found.Results[0]
parts := place.Components

fmt.Println(mygeocode.StringValue(place.Formatted))   // Brandenburg Gate, State of Berlin, Germany
fmt.Println(mygeocode.StringValue(place.Precision))   // house
fmt.Println(mygeocode.StringValue(parts.POI))         // Brandenburg Gate
fmt.Println(mygeocode.StringValue(parts.CountryCode)) // de
fmt.Println(parts.City == nil)                        // true (not every address has every part)
```

Every text, number and flag is a pointer (`*string`, `*float64`, `*int64`, `*bool`), so a field the API did not send
is `nil`, never a made-up `0` or `""`. `mygeocode.StringValue`, `FloatValue`, `IntValue` and `BoolValue` read one with
a zero default. The fields this SDK reads are never renamed or removed from the API: if one ever has to go, it keeps
coming back as `nil`, and your code keeps working. Lists are slices, empty when the API sent none, and the address
parts in `Components` are always text. A field of an unexpected type reads as `nil` instead of failing the call.

Every result also has `Raw`, the JSON object exactly as the API sent it (a `json.RawMessage`), so fields added to the
API later are readable before the SDK knows them: `json.Unmarshal(place.Raw, &yourStruct)`.

What each method returns, and its fields:

| Type | Returned by | Fields |
|---|---|---|
| `ForwardResponse` | `Forward` | `Status`, `Query`, `Results` (`[]Place`), `Meta`, `Raw` |
| `ReverseResponse` | `Reverse`, each entry of `ReverseBatch` | `Status`, `Lat`, `Lon`, `Result` (`*Place`), `Boundaries`, `Meta`, `Raw` |
| `Place` | inside forward and reverse results | `Formatted`, `Lat`, `Lon`, `Type`, `Precision`, `Confidence`, `PlaceID`, `Components`, `Bounds`, `DistanceM`, `FormattedPlace`, `Place`, `Parcel`, `Codes`, `Population`, `UNLocode`, `Timezone`, `ElevationM`, `Wikidata`, `CountryInfo`, `Sun`, `Regions`, `Districts`, `Census`, `Demographics`, `Road`, `Intersection`, `Site`, `Raw` |
| `Components` | `Place.Components` | `POI`, `Building`, `Unit`, `Level`, `HouseNumber`, `Road`, `Lot`, `Suburb`, `Municipality`, `City`, `County`, `State`, `StateCode`, `Postcode`, `Country`, `CountryCode`, `OfficialName`, `Raw` |
| `Bounds` | `Place.Bounds` | `North`, `South`, `East`, `West`, `Raw` |
| `AutocompleteResponse` | `Autocomplete` | `Status`, `Query`, `Suggestions` (`[]Suggestion`), `Meta`, `Raw` |
| `Suggestion` | inside autocomplete results | `Text`, `Lat`, `Lon`, `Type`, `Precision`, `PlaceID`, `Raw` |
| `IPInfo` | `IP`, `IPv4`, `IPv6`, each entry of the IP batches | `Status`, `IP`, `Version`, `Found`, `Prefix`, `Country`, `CountryCode`, `Continent`, `ContinentCode`, `Region`, `RegionCode`, `City`, `Postcode`, `Lat`, `Lon`, `ElevationM`, `Timezone`, `UTCOffset`, `UTCOffsetSeconds`, `IsEU`, `ASN`, `ASName`, `Org`, `UsageType`, `Registry`, `Allocated`, `Tags`, `IsAnycast`, `IsSatellite`, `IsDatacenter`, `IsProxy`, `IsVPNProvider`, `IsTor`, `Crawler`, `IsPrivate`, `Reason`, `AbuseEmail`, `Threat`, `UpdatedAt`, `Error` (`*ErrorDetail`), `Meta`, `Raw` |
| `TimezoneInfo` | `Timezone`, each entry of `TimezoneBatch` | `Status`, `Lat`, `Lon`, `Timezone`, `Name`, `Abbreviation`, `UTCOffset`, `UTCOffsetSeconds`, `RawOffsetSeconds`, `DST`, `DSTOffsetSeconds`, `Timestamp`, `LocalTime`, `CountryCode`, `IsOcean`, `WindowsTimezone`, `NextTransition` (`*TimezoneTransition`), `Meta`, `Raw` |
| `TimezoneTransition` | `TimezoneInfo.NextTransition` | `Timestamp`, `UTCOffset`, `UTCOffsetSeconds`, `DST`, `Abbreviation`, `Raw` |
| `ElevationResponse` | `Elevation`, `ElevationBatch` | `Status`, `Results` (`[]ElevationPoint`), `Meta`, `Raw` |
| `ElevationPoint` | inside elevation results | `Lat`, `Lon`, `ElevationM`, `ResolutionM`, `Raw` |
| `PostcodeResponse` | `Postcode` | `Status`, `Postcode`, `CountryCode`, `Results` (`[]PostcodePlace`), `Meta`, `Raw` |
| `PostcodePlace` | inside postcode results | `Postcode`, `Place`, `Region`, `RegionCode`, `Country`, `CountryCode`, `Lat`, `Lon`, `Raw` |
| `ForwardBatchResponse` | `ForwardBatch` | `Status`, `Results` (`[]ForwardBatchEntry`), `Meta`, `Raw` |
| `ForwardBatchEntry` | inside forward batch results | `Query`, `Results` (`[]Place`), `Raw` |
| `ReverseBatchResponse` | `ReverseBatch` | `Status`, `Results` (`[]ReverseResponse`), `Meta`, `Raw` |
| `IPBatchResponse` | `IPBatch`, `IPv4Batch`, `IPv6Batch` | `Status`, `Results` (`[]IPInfo`), `Meta`, `Raw` |
| `TimezoneBatchResponse` | `TimezoneBatch` | `Status`, `Timestamp`, `Results` (`[]TimezoneInfo`), `Meta`, `Raw` |

The extra blocks you ask for with the flags (`Timezone`, `Elevation`, `Wikidata` and the rest) are `json.RawMessage`
fields of `Place`, `nil` when not asked for. On an entry inside a batch, `Meta` is empty; the batch response has the
meta.

Every response has `Meta`, filled from the response headers (`nil`, or `""` for `RequestID`, when a header was not
sent):

| Field | Header | Meaning |
|---|---|---|
| `RequestID` | `X-Request-Id` | Quote it when you contact support. |
| `QuotaLimit` | `X-Quota-Limit` | Free requests a day for this key's account or network. |
| `QuotaUsed` | `X-Quota-Used` | Free requests used today. |
| `QuotaFreeRemaining` | `X-Quota-Free-Remaining` | Free requests left today. |
| `QuotaReset` | `X-Quota-Reset` | Unix time when the free requests reset. |
| `CreditsRemaining` | `X-Credits-Remaining` | Prepaid credit left, when a key with credit is used. |

## Errors

Every error a method returns implements `mygeocode.MyGeocodeError`. Reach the concrete type with `errors.As`:

| Type | When |
|---|---|
| `*ApiError` | The API answered with an error. It has `Status` (HTTP), `Code` (the API's error code, such as `invalid_request`, `invalid_key` or `quota_exceeded`), `Message`, `Param` (the parameter at fault), `RequestID`, `RetryAfter` (a `*time.Duration`, `nil` when the API asked for no wait) and `Body` (the answer as sent). |
| `*RequestTimeoutError` | No answer within the timeout. |
| `*NetworkError` | The connection failed: DNS, refused, reset or TLS, or your context was cancelled (`errors.Is(err, context.Canceled)` holds). The original error is in `Err` and `Unwrap`. |
| `*DecodeError` | The answer could not be read as a JSON object. |

```go
found, err := c.Forward(ctx, "Brandenburg Gate, Berlin", nil)
var apiErr *mygeocode.ApiError
switch {
case errors.As(err, &apiErr):
	fmt.Println(apiErr.Status, apiErr.Code, apiErr.Message)
case err != nil:
	fmt.Println("could not reach the API:", err)
default:
	fmt.Println(len(found.Results), "results") // 5 results
}
```

The [errors page](https://www.mygeocode.com/docs/errors/) lists every error code.

## Retries

The client retries on its own, up to `WithMaxRetries` times (default 2):

- **429** (too many requests): waits for `Retry-After`, or until `X-Quota-Reset`. If that is longer than
  `WithMaxRetryWait` (default 60 seconds), for example when the day's free requests are gone, it returns the error
  straight away instead of waiting.
- **500**: tries once more after one second.
- **503**: waits for `Retry-After`, or 1, 2 and then 4 seconds.
- **Timeouts and connection errors**: retried with the same 1, 2, 4 second pattern for GET requests. The POST batch
  calls are not retried after a connection error, so a batch is never charged twice.

Other errors are returned at once. Waits end early when your context ends.

## Configuration

All options are optional:

```go
c := mygeocode.NewClient(
	mygeocode.WithKey("YOUR_KEY"),                          // default: MYGEOCODE_KEY, else no key
	mygeocode.WithBaseURL("https://api.mygeocode.com/v1"),
	mygeocode.WithTimeout(10*time.Second),                  // for each attempt, from sending to the last byte
	mygeocode.WithMaxRetries(2),
	mygeocode.WithMaxRetryWait(60*time.Second),
	mygeocode.WithLang("en"),
)
```

A deadline or cancel on the context you pass applies to the whole call, retries included; the timeout applies to each
attempt. `WithTransport`, `WithSleep` and `WithClock` replace the HTTP transport, the wait and the clock the retry logic
uses; they are meant for tests, proxies and custom TLS settings, see below.

The default transport is Go's own (`http.DefaultTransport` settings, through a fresh `http.Client`): it uses the
system's certificate store and the `HTTPS_PROXY` and `NO_PROXY` environment variables, and reuses connections.

## Testing your code (transport hook)

`WithTransport` takes any `http.RoundTripper`. Give it your own to test code that uses the SDK without touching the
network:

```go
type canned struct{}

func (canned) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	w.WriteString(`{"status":"ok","lat":48.8584,"lon":2.2945,"result":{"formatted":"5 Avenue Anatole France, 75007 Paris, France","lat":48.858821,"lon":2.294596}}`)
	return w.Result(), nil
}

c := mygeocode.NewClient(
	mygeocode.WithTransport(canned{}),
	mygeocode.WithSleep(func(ctx context.Context, d time.Duration) error { return nil }), // no real waiting in retry tests
)
address, err := c.Reverse(ctx, 48.8584, 2.2945, nil)
fmt.Println(mygeocode.StringValue(address.Result.Formatted)) // 5 Avenue Anatole France, 75007 Paris, France
```

The same hook sets a proxy or your own TLS settings: `WithTransport(&http.Transport{Proxy: ..., TLSClientConfig:
...})`. The client still adds its own headers, timeout and retries.

### Running the SDK's own tests

The tests use only the standard library. They answer from real recorded API responses (and a small local server for
the transport) and never call the API:

```bash
go test ./...
```

## Requirements

Go 1.13 or newer, nothing else. Tested on every Go release from 1.13 to 1.27, with `go vet` on the oldest and the
newest, and the race detector.

## Links

- [API documentation](https://www.mygeocode.com/docs/)
- [Pricing and free tier](https://www.mygeocode.com/pricing/)
- [Dashboard and API keys](https://www.mygeocode.com/dashboard)
- [Package documentation](https://pkg.go.dev/github.com/mygeocode/mygeocode-go)
- [Issues](https://github.com/mygeocode/mygeocode-go/issues)
- Support: [support@mygeocode.com](mailto:support@mygeocode.com)

## Licence

MIT. See [LICENSE](LICENSE).
