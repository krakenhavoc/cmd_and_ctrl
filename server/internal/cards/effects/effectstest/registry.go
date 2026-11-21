package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// registerForTest registers spec and schedules its removal from the
// registry when the test completes. Register's map write is the only
// state a Spec leaves behind that a later pass of the same test can
// collide with (#1070): init()-time registrations run once no matter
// how many times `go test -count=N` re-executes the binary, but a
// Register call inside a test function body runs on every pass, and
// the package-level registry does not reset between them. A test that
// registers a fixed OracleID through this helper instead of calling
// Register directly cleans up after itself, so `-count=N` sees the
// same empty starting registry on every pass. See registry.go's
// package doc, which already names this helper.
func RegisterForTest(t *testing.T, spec effects.Spec) {
	t.Helper()
	effects.Register(spec)
	t.Cleanup(func() { delete(effects.Registry, spec.OracleID) })
}
