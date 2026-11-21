package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vulpikeet — Creature — Fox Bird {3}{W}, 2/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"Mutate {2}{W} (If you cast this spell for its mutate cost, put it
//	 over or under target non-Human creature you own. They mutate into
//	 the creature on top plus all abilities from under it.)
//	 Flying
//	 Whenever this creature mutates, put a +1/+1 counter on it."
//
// Mutate is the engine's (ADR 0223), as on Everquill Phoenix: the
// trigger fires on the merge whichever card ends on top (CR 702.140d,
// 702.140e), and not when the card is cast for its mana cost.
// The counters go on the merged permanent, "it" (CR 730.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "11d0c8a4-1dd0-4707-9c1d-19ed27d1934e",
		Name:             "Vulpikeet",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{2}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventMutate, effects.ThisCreatureMutates, "Vulpikeet — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
