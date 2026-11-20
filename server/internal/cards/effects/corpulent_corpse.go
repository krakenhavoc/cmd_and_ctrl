package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Corpulent Corpse — Creature — Zombie {5}{B}, 3/3 (Time Spiral, #5596):
//
//	"Fear (This creature can't be blocked except by artifact creatures
//	 and/or black creatures.)
//	 Suspend 5—{B} (Rather than cast this card from your hand, you may
//	 pay {B} and exile it with five time counters on it. At the
//	 beginning of your upkeep, remove a time counter. When the last is
//	 removed, you may cast it without paying its mana cost. It has
//	 haste.)"
//
// A 3/3 with fear (CR 702.36b) whose only other text is suspend (CR
// 702.62a). The keyword is the engine's: the special action, the time
// counters, the upkeep countdown, the free cast when the last counter
// comes off and the haste it then has are all grown from the one
// declaration.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b3a84b75-a6a2-4521-a9f1-92596796c703",
		Name:            "Corpulent Corpse",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"fear"},
		SpecialActions: []game.SpecialAction{
			Suspend(5, "{B}"),
		},
	})
}
