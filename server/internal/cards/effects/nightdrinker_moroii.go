package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nightdrinker Moroii — Creature — Vampire {3}{B}, 4/2 (Murders at
// Karlov Manor, #4469):
//
//	"Flying
//	 When this creature enters, you lose 3 life.
//	 Disguise {B}{B} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)"
//
// Cast face down, it enters with no text (CR 708.3), so the life loss
// never happens, and turning it face up does not trigger it either
// (CR 708.8): that is the disguise's point. Flying rides
// PrintedKeywords, which a face-down permanent does not have.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "07d830a7-f0a2-4d46-9a53-101dca681603",
		Name:             "Nightdrinker Moroii",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Disguise("{B}{B}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Nightdrinker Moroii — you lose 3 life", Do(GainLife{Amount: -3})),
		},
	})
}
