package deck

import "testing"

// TestingSetMoxfieldAPIHost points the Moxfield fetcher at `host`
// for the duration of t, restoring the original on cleanup. Exported
// so tests in other packages (e.g. the lobby HTTP integration tests)
// can stand up an httptest.Server without reaching into this
// package's unexported state.
//
// The function takes a *testing.T rather than testing.TB so an
// errant production caller is more obviously wrong (there's no
// legitimate non-test use).
func TestingSetMoxfieldAPIHost(t *testing.T, host string) {
	t.Helper()
	orig := moxfieldAPIHost
	moxfieldAPIHost = host
	t.Cleanup(func() { moxfieldAPIHost = orig })
}

// TestingSetArchidektAPIHost is the Archidekt counterpart. See
// TestingSetMoxfieldAPIHost.
func TestingSetArchidektAPIHost(t *testing.T, host string) {
	t.Helper()
	orig := archidektAPIHost
	archidektAPIHost = host
	t.Cleanup(func() { archidektAPIHost = orig })
}
