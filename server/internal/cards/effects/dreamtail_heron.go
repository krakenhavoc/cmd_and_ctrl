package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreamtail Heron — Creature — Elemental Bird {4}{U}, 3/4 (Ikoria: Lair of Behemoths, #5512):
//
//	"Mutate {3}{U} (If you cast this spell for its mutate cost, put it
//	 over or under target non-Human creature you own. They mutate into
//	 the creature on top plus all abilities from under it.)
//	 Flying
//	 Whenever this creature mutates, draw a card."
//
// Mutate is the engine's (ADR 0223), as on Everquill Phoenix: the
// trigger fires on the merge whichever card ends on top (CR 702.140d,
// 702.140e), and not when the card is cast for its mana cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "1cadb1e1-19b6-4ce5-8eb5-7851cec7dfa3",
		Name:             "Dreamtail Heron",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Mutate("{3}{U}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventMutate, ThisCreatureMutates, "Dreamtail Heron — draw a card",
				Do(DrawCards{N: 1})),
		},
	})
}
