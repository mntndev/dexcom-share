package dexcomshare_test

import (
	"context"
	"fmt"
	"log"

	dexcomshare "github.com/mntndev/dexcom-share"
)

// This example is compiled but not run, since it talks to the real Share API.
func Example() {
	ctx := context.Background()

	client, err := dexcomshare.NewClient(ctx, "username", "password")
	if err != nil {
		log.Fatal(err)
	}

	// The ten most recent readings from the last hour.
	entries, err := client.ReadGlucose(ctx, 60, 10)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		fmt.Printf("%s %d mg/dL %s %s\n",
			entry.WT.Local().Format("15:04"),
			entry.Value,
			entry.Trend.Arrow(),
			entry.Trend.Description(),
		)
	}
}

// Accounts registered outside the United States live on a different host.
func ExampleWithBaseURL() {
	_, err := dexcomshare.NewClient(
		context.Background(),
		"username",
		"password",
		dexcomshare.WithBaseURL(dexcomshare.BaseURLOutsideUS),
	)
	if err != nil {
		log.Fatal(err)
	}
}
