// Smoke test against the live API: 8 calls, one per endpoint family. Run it by hand, never from the unit tests:
//
//	go run ./examples/smoke
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mygeocode/mygeocode-go"
)

func main() {
	ctx := context.Background()
	c := mygeocode.NewClient()
	failed := false
	check := func(name string, err error, detail func() string) {
		if err != nil {
			failed = true
			fmt.Printf("FAIL %-15s %v\n", name, err)
			return
		}
		fmt.Printf("ok   %-15s %s\n", name, detail())
	}
	s := mygeocode.StringValue

	f, err := c.Forward(ctx, "Brandenburg Gate, Berlin", &mygeocode.ForwardParams{Limit: 1})
	check("forward", err, func() string { return s(f.Results[0].Formatted) })

	r, err := c.Reverse(ctx, 48.8584, 2.2945, nil)
	check("reverse", err, func() string { return s(r.Result.Formatted) })

	a, err := c.Autocomplete(ctx, "1600 Pennsylvania", &mygeocode.AutocompleteParams{Limit: 2})
	check("autocomplete", err, func() string { return s(a.Suggestions[0].Text) })

	ip, err := c.IPv4(ctx, "8.8.8.8", nil)
	check("ipv4", err, func() string { return s(ip.ASName) })

	tz, err := c.Timezone(ctx, 35.6762, 139.6503, nil)
	check("timezone", err, func() string { return s(tz.Timezone) })

	el, err := c.Elevation(ctx, 27.9881, 86.9250, nil)
	check("elevation", err, func() string { return fmt.Sprint(mygeocode.FloatValue(el.Results[0].ElevationM)) })

	pc, err := c.Postcode(ctx, "10001", &mygeocode.PostcodeParams{Country: "US"})
	check("postcode", err, func() string { return s(pc.Results[0].Place) })

	eb, err := c.ElevationBatch(ctx, []mygeocode.Point{{Lat: 48.8584, Lon: 2.2945}, {Lat: -33.8568, Lon: 151.2153}}, nil)
	check("elevation batch", err, func() string { return fmt.Sprintf("%d points", len(eb.Results)) })

	if failed {
		os.Exit(1)
	}
}
