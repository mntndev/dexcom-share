# dexcom-share

[![CI](https://github.com/mntndev/dexcom-share/actions/workflows/ci.yml/badge.svg)](https://github.com/mntndev/dexcom-share/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/mntndev/dexcom-share.svg)](https://pkg.go.dev/github.com/mntndev/dexcom-share)

`dexcom-share` is a Go client for the undocumented Dexcom Share API. Use it to read
near-realtime glucose data uploaded from a Dexcom continuous glucose monitor.

The Share API is not public, not documented, and not supported by Dexcom. It can
change or disappear without notice. You have been warned.

## Install

```sh
go get github.com/mntndev/dexcom-share
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	dexcomshare "github.com/mntndev/dexcom-share"
)

func main() {
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
```

`ReadGlucose` takes a lookback window in minutes (1–1440) and a maximum number of
entries (1–288), and returns the most recent readings first.

### Readings

`GlucoseEntry.Value` is in mg/dL; call `MmolL` for mmol/L. `Trend` is a typed
string with `Arrow`, `Description`, and `Known` helpers. `DT`, `WT`, and `ST` are
parsed timestamps that embed `time.Time`, so the usual time methods work directly.
`DT` carries the offset the device reported; `WT` and `ST` are UTC.

### Accounts outside the United States

Dexcom serves non-US accounts from a different host:

```go
client, err := dexcomshare.NewClient(ctx, "username", "password",
	dexcomshare.WithBaseURL(dexcomshare.BaseURLOutsideUS),
)
```

Use `WithHTTPClient` to supply your own `*http.Client` for timeouts, proxies, or
instrumentation.

### Errors

Failures wrap a sentinel you can match with `errors.Is` — `ErrAuthenticationFailed`,
`ErrLoginFailed`, `ErrReadGlucoseFailed`, `ErrInvalidMinutes`, `ErrInvalidMaxCount` —
and, when the API reported a reason, an `*APIError` you can reach with `errors.As`:

```go
entries, err := client.ReadGlucose(ctx, 60, 10)

var apiErr *dexcomshare.APIError
if errors.As(err, &apiErr) && apiErr.Code == "SessionIdNotFound" {
	// The session expired; build a new client.
}
```

Sessions do not refresh themselves. When one expires, create a new client.

## Development

```sh
go test -race -coverprofile=coverage.out ./...
```

Tests run against `httptest` stubs plus one session recorded from the real API
with [go-vcr](https://github.com/dnaeon/go-vcr) (`testdata/TestRecordedSession.yaml`).
No credentials are needed, and the recorded cassette has them redacted.

## License

MIT
