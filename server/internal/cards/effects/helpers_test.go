package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestIsBasicLandFindsSnowBasics is the #1334 regression: a snow
// basic's type line is "Basic Snow Land — Plains", where the words
// "Basic" and "Land" never appear adjacent, so the old
// containsFoldASCII(typeLine, "basic land") body missed it. The fix
// reads the Basic supertype instead (CR 205.4a), which is what
// b30IsBasicLandCard already did — see the alias comment on that
// function.
func TestIsBasicLandFindsSnowBasics(t *testing.T) {
	cases := []struct {
		name     string
		typeLine string
		want     bool
	}{
		{"ordinary basic", "Basic Land — Forest", true},
		{"snow basic", "Basic Snow Land — Plains", true},
		{"nonbasic land", "Land — Gate", false},
		{"nonbasic snow land", "Snow Land — Gate", false},
		{"nonbasic dual", "Land — Island Swamp", false},
		{"not a land at all", "Creature — Human Wizard", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := game.Card{Name: tc.name, TypeLine: tc.typeLine}
			if got := IsBasicLand(c); got != tc.want {
				t.Errorf("IsBasicLand(%q) = %v, want %v", tc.typeLine, got, tc.want)
			}
		})
	}
}
