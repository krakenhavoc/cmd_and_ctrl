package botarena_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
)

// #5083: `heuristic-sourcecounters` is today's heuristic with the price
// of a row's +1/+1 counter on its own source on, an arena name only.
func TestSourceCountersContestant(t *testing.T) {
	const in = "heuristic-sourcecounters"
	if botarena.SourceCountersContestant != in {
		t.Fatalf("SourceCountersContestant is %q, want %q", botarena.SourceCountersContestant, in)
	}
	got, err := botarena.ParseContestant(in)
	if err != nil {
		t.Fatalf("ParseContestant(%q): %v", in, err)
	}
	if want := (botarena.SeatSpec{Tier: tiers.Heuristic, Variant: botarena.VariantSourceCounters}); got != want {
		t.Errorf("ParseContestant(%q) = %+v, want %+v", in, got, want)
	}
	if got.Contestant() != in || got.Label() != in {
		t.Errorf("%q round-trips as contestant %q, label %q", in, got.Contestant(), got.Label())
	}
	if _, err := tiers.Parse(in); err == nil {
		t.Fatalf("tiers.Parse accepts %q; it would be offered in the lobby", in)
	}
}
