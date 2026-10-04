package mcpseat

import (
	"fmt"
	"strings"
)

// Result is what every tool handler returns: the text the model reads,
// and whether the call failed. transport.go converts it to the SDK's
// result type, and nothing else in the package knows that type exists.
type Result struct {
	Text    string
	IsError bool
}

// textResult builds a successful result from lines.
func textResult(lines ...string) Result {
	return Result{Text: strings.Join(lines, "\n")}
}

// errorResult is a tool-level failure: the model sees it and can
// correct course, which a protocol error would not let it do.
func errorResult(format string, args ...any) Result {
	return Result{Text: fmt.Sprintf(format, args...), IsError: true}
}
