package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phantasmal Shieldback — Creature — Turtle Illusion {U}, 1/3
// (Foundations Jumpstart, #5464):
//
//	"When this creature becomes the target of a spell or ability,
//	 sacrifice it.
//	 When this creature dies, draw a card."
//
// The Illusion clause is Phantasmal Image's: a trigger on the Shieldback
// being targeted by any spell or ability, whose resolution sacrifices it
// if it is still the same permanent. The sacrifice is a death, so the
// dies trigger draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e514bc9-7c8c-4257-9dbc-19a78f61df05",
		Name:         "Phantasmal Shieldback",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.Self,
				"Phantasmal Shieldback — sacrifice it",
				effects.SacrificeThisIfStillOnBattlefield),
			effects.WhenThisDies("Phantasmal Shieldback — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
