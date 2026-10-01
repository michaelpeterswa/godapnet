# godapnet

A Go library for sending messages to amateur radio POCSAG pagers over the
[DAPNET](https://hampager.de) network.

```sh
go get github.com/michaelpeterswa/godapnet
```

## Usage

```go
package main

import (
	"context"
	"log"

	"github.com/michaelpeterswa/godapnet"
)

func main() {
	// Posts to godapnet.DAPNetURL with a 30 second timeout unless overridden
	// with godapnet.WithURL or godapnet.WithHTTPClient.
	sender := godapnet.NewSender("x1xxx", "password")

	config := godapnet.NewMessageConfig(
		[]string{"x1xxx"}, // destination callsigns
		godapnet.WithTransmitterGroups("us-wa"),
		godapnet.WithPrefix("x1xxx"), // each page reads "x1xxx: <text>"
	)

	if err := sender.Send(context.Background(), "hello from godapnet", config); err != nil {
		log.Fatal(err)
	}
}
```

Text longer than a page is split between words into several pages, each
carrying the prefix. Pages are 80 characters
(`godapnet.Alphapoc602RMaxMessageLength`) unless set with
`WithMaxMessageLength`, and are sent first page first unless
`WithReverseOrder` is given. `WithEmergency` marks the call as an emergency.

`Send` checks the config before sending anything and returns `ErrNoCallsigns`,
`ErrInvalidMaxLength`, `ErrPrefixTooLong` or `ErrEmptyText` for a message it
cannot send. If DAPNET rejects a page, the error wraps a `*godapnet.StatusError`
holding the status code and the start of the response body:

```go
var statusErr *godapnet.StatusError
if errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusUnauthorized {
	// bad credentials
}
```

## Development

```sh
make        # wire up git hooks and install commitlint
make test   # go test -race -shuffle=on ./...
make lint   # golangci-lint and yamllint, the same checks CI runs
```

Commits follow [Conventional Commits](https://www.conventionalcommits.org/);
semantic-release cuts a GitHub release from them on every push to `main`.

## License

MIT
