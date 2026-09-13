package aiseat_test

import (
	"os"
	"testing"
)

// gametest_test.go — the opt-in gate for whole-game tests.
//
// Most tests in this package are ordinary unit tests: a runner, a
// policy, a deadline. A handful instead play COMPLETE four-seat games
// to a winner — the random fuzzer, the heuristic head-to-head, the
// Layer A absorption census, the model funnel's outage drill. They are
// the most valuable tests here; between them they caught four engine
// bugs on 2026-09-11 that no unit test was looking for.
//
// They are also, together, most of this package's runtime, and the
// package is the slowest thing in CI. Ungated the whole package runs
// in ~45s; under `-race` on a shared self-hosted runner with a dozen
// jobs in flight it crossed Go's 600s per-package default and failed
// `main` plus every open PR (fixed by #471, which raised the cap to
// 30m — headroom, not a licence).
//
// So: skipped by default, run in full on the nightly. The trade is
// deliberate and worth stating plainly — a bug these games would catch
// now surfaces the next morning rather than in the PR that caused it.
// That is the same bargain e2e-nightly.yml already struck, for the same
// reason, and it is the right one only while the nightly actually runs
// them. If the nightly stops running, this gate becomes a way of
// deleting the tests without noticing.
//
// Run them locally exactly as the nightly does:
//
//	AISEAT_GAME_TESTS=1 go test ./internal/aiseat/... -race -timeout 30m
//
// The per-test count knobs (AISEAT_H2H_GAMES, AISEAT_FUNNEL_GAMES,
// AISEAT_SOAK_GAMES, …) still work and are independent of this gate:
// this one decides IF the game tests run, those decide HOW MANY.
func requireGameTests(t *testing.T) {
	t.Helper()
	if os.Getenv("AISEAT_GAME_TESTS") == "" {
		t.Skip("whole-game test: set AISEAT_GAME_TESTS=1 to run (the nightly does)")
	}
}
