package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keldon Halberdier — Creature — Human Warrior {4}{R}, 4/1 (Time Spiral, #5596):
//
//	"First strike
//	 Suspend 4—{R} (Rather than cast this card from your hand, you may
//	 pay {R} and exile it with four time counters on it. At the
//	 beginning of your upkeep, remove a time counter. When the last is
//	 removed, you may cast it without paying its mana cost. It has
//	 haste.)"
//
// A 4/1 first striker whose only other text is suspend (CR 702.62a). The
// keyword is the engine's: the special action, the time counters, the
// upkeep countdown, the free cast when the last counter comes off and the
// haste it then has are all grown from the one declaration.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a1472a2d-e6f7-4d65-8bb2-4ecfae2d2764",
		Name:            "Keldon Halberdier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		SpecialActions: []game.SpecialAction{
			Suspend(4, "{R}"),
		},
	})
}
