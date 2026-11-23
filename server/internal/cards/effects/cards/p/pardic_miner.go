package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pardic Miner — Creature — Dwarf {1}{R}:
//
//	"Sacrifice this creature: Target player can't play lands this turn."
//
// ADR 0109 §4 (#1895). The sacrifice is the cost, so the creature is
// gone when the ability resolves; the ban is a stored ModCantPlayLands
// record on the player (swept at cleanup, CR 514.2), not a static, which
// is what lets it outlive its source. The record names the Miner as its
// source from last-known information, so the player's refusal says
// "You can't play lands this turn — Pardic Miner".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1788b54-4cd9-459f-b2c9-06773e68e9e3",
		Name:         "Pardic Miner",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Target player can't play lands this turn.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CantPlayLandsThisTurn{}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
