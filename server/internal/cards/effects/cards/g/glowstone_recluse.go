package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glowstone Recluse — Creature — Spider {2}{G}, 2/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"Mutate {3}{G} (If you cast this spell for its mutate cost, put it
//	 over or under target non-Human creature you own. They mutate into
//	 the creature on top plus all abilities from under it.)
//	 Reach
//	 Whenever this creature mutates, put two +1/+1 counters on it."
//
// Mutate is the engine's (ADR 0223), as on Everquill Phoenix: the
// trigger fires on the merge whichever card ends on top (CR 702.140d,
// 702.140e), and not when the card is cast for its mana cost.
// The counters go on the merged permanent, "it" (CR 730.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "b2b76070-f103-4f83-ad21-119e6352f2cd",
		Name:             "Glowstone Recluse",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"reach"},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{3}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventMutate, effects.ThisCreatureMutates, "Glowstone Recluse — put two +1/+1 counters on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
	})
}
