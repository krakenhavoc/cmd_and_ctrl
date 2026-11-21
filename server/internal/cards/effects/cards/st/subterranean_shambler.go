package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Subterranean Shambler — Creature — Elemental, {3}{R}, 2/3:
//
//	"Echo {3}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters or leaves the battlefield, it deals 1 damage to each creature without flying."
//
// "Enters or leaves the battlefield" is one ability with two trigger
// conditions; the damage goes to each creature without flying as it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "d06be0a6-0d45-4e95-b132-88ce80dd7cb6",
		Name:     "Subterranean Shambler",
		// ADR 0126 §6: a creature with flying is spared.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Echo("Subterranean Shambler", "{3}{R}"),
			effects.WhenThisEntersOrLeaves(
				"Subterranean Shambler — 1 damage to each creature without flying",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.WithoutKeyword("flying")), 1)
				}),
		},
	})
}
