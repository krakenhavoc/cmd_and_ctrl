package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quicksilver, Speedster — Legendary Creature — Mutant Hero {3}{U}{R},
// 3/4 (Marvel Super Heroes Commander, #4455):
//
//	"Flash
//	 Double strike, haste
//	 As long as Quicksilver is tapped, you may cast spells as though they
//	 had flash."
//
// The keywords are printed. The last line is Captain Mar-Vell, Space-Born's
// derived cast-timing declaration (Spec.CastTimings) gated on the
// permanent's own tapped status (CastTimingRule.SourceTapped) instead of on
// an opponent's spell: read live on every query, so an attacking Quicksilver
// opens the window, an untapped one shuts it, and one that loses its
// abilities or leaves says nothing. It is about casting, so it does not let
// its controller play a land at instant speed.
//
// No simplification.
func init() {
	const label = "As long as Quicksilver is tapped, you may cast spells as though they had flash."
	Register(Spec{
		OracleID:        "cc20f910-03e9-4019-a62a-2e713ddb3c06",
		Name:            "Quicksilver, Speedster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "double strike", "haste"},
		CastTimings:     []game.CastTimingRule{CastAsThoughFlashWhileSourceTapped(label)},
	})
}
