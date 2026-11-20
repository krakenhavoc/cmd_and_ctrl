package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Majestic Auricorn — Creature — Unicorn {4}{W}, 4/4 (Ikoria: Lair of Behemoths, #5512):
//
//	"Mutate {3}{W} (If you cast this spell for its mutate cost, put it
//	 over or under target non-Human creature you own. They mutate into
//	 the creature on top plus all abilities from under it.)
//	 Vigilance
//	 Whenever this creature mutates, you gain 4 life."
//
// Mutate is the engine's (ADR 0223), as on Everquill Phoenix: the
// trigger fires on the merge whichever card ends on top (CR 702.140d,
// 702.140e), and not when the card is cast for its mana cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "9c13f77a-800c-4fb6-b54a-89a96e12d279",
		Name:             "Majestic Auricorn",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"vigilance"},
		AlternativeCosts: []game.AlternativeCost{Mutate("{3}{W}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventMutate, ThisCreatureMutates, "Majestic Auricorn — you gain 4 life",
				Do(GainLife{Amount: 4})),
		},
	})
}
