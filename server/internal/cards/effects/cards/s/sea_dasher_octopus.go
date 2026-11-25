package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sea-Dasher Octopus — Creature — Octopus {1}{U}{U}, 2/2 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Mutate {1}{U} (If you cast this spell for its mutate cost, put it over or
//	 under target non-Human creature you own. They mutate into the creature on
//	 top plus all abilities from under it.)
//	 Flash
//	 Whenever this creature deals combat damage to a player, draw a card."
//
// Mutate is the engine's (ADR 0223). The merged creature has every component's
// abilities (CR 702.140e), so flash is a printed keyword of the card and the
// draw trigger works from under another creature too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "0b718d21-6873-4613-ad15-660d30527fcd",
		Name:             "Sea-Dasher Octopus",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flash"},
		Purpose:          game.Purpose{Draws: 1},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Sea-Dasher Octopus — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
