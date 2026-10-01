package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demon of Death's Gate — Creature — Demon {6}{B}{B}{B}, 9/9:
//
//	"You may pay 6 life and sacrifice three black creatures rather
//	 than pay this spell's mana cost.
//	 Flying, trample"
//
// A #1727 proof card: a life component and a sacrifice component on
// ONE alternative cost, both validated before either is paid. A player
// at 5 life is refused the cast with all three creatures still on the
// battlefield, and one who names a green creature is refused with
// their life untouched.
func init() {
	Register(Spec{
		OracleID:        "be465805-1a37-4212-acf1-5dfec890b338",
		Name:            "Demon of Death's Gate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		AlternativeCosts: []game.AlternativeCost{
			SacrificeInstead(3, "three black creatures", 6, OfColor("B"), Creature()),
		},
	})
}
