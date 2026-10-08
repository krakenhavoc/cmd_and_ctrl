package effects

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lands_untapped_test.go — ADR 0136 §2, owner answer 4: a ramp spell
// says how many of its lands enter untapped (Purpose.LandsUntapped),
// and Register refuses a count it cannot hold.

func TestRegisterRefusesMoreLandsUntappedThanLands(t *testing.T) {
	for _, c := range []struct {
		name string
		p    game.Purpose
		want string
	}{
		{"more than lands", game.Purpose{Lands: 1, LandsUntapped: 2}, "more lands enter untapped"},
		{"no lands", game.Purpose{LandsUntapped: 1}, "more lands enter untapped"},
		{"negative", game.Purpose{Lands: 1, LandsUntapped: -1}, "negative amount"},
	} {
		spec := Spec{OracleID: "test-0136-lands-untapped-" + c.name, Name: "Test " + c.name, Purpose: c.p}
		if msg := registerPanics(spec); !strings.Contains(msg, c.want) {
			t.Errorf("%s: panic %q, want one mentioning %q", c.name, msg, c.want)
		}
	}
}

// The three curated spells whose lands enter untapped declare it.
func TestCuratedRampDeclaresLandsUntapped(t *testing.T) {
	for _, c := range []struct {
		name, oracle string
		lands        int
	}{
		{"Harrow", "705509e9-a034-4a5a-9c65-66f58748b8a2", 2},
		{"Nature's Lore", "78826359-fe63-44ad-adc4-a17ffcd710e4", 1},
		{"Three Visits", "1b882a0e-0ede-4d1a-bd1a-9b7cffbcde8e", 1},
	} {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Fatalf("%s is not registered", c.name)
		}
		if spec.Purpose.Lands != c.lands || spec.Purpose.LandsUntapped != c.lands {
			t.Errorf("%s declares %+v, want %d lands, all untapped", c.name, spec.Purpose, c.lands)
		}
	}
}
