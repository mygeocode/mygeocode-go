// Quickstart: one forward and one reverse call, without a key.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mygeocode/mygeocode-go"
)

func main() {
	ctx := context.Background()
	c := mygeocode.NewClient() // no key: the network's free requests apply

	found, err := c.Forward(ctx, "Brandenburg Gate, Berlin", &mygeocode.ForwardParams{Limit: 1})
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range found.Results {
		fmt.Printf("%s (%v, %v)\n", mygeocode.StringValue(p.Formatted), mygeocode.FloatValue(p.Lat), mygeocode.FloatValue(p.Lon))
	}

	address, err := c.Reverse(ctx, 48.8584, 2.2945, nil)
	if err != nil {
		log.Fatal(err)
	}
	if address.Result != nil {
		fmt.Println(mygeocode.StringValue(address.Result.Formatted))
	}
	if left := address.Meta.QuotaFreeRemaining; left != nil {
		fmt.Println("free requests left today:", *left)
	}
}
