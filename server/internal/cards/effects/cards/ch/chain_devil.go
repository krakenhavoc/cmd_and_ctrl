package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chain Devil — Creature — Devil {3}{B}, 4/2 (Commander Legends: Battle for
// Baldur's Gate, #5948):
//
//	"Animate Chains — When this creature enters, each player sacrifices a
//	 nontoken creature of their choice."
//
// Accursed Marauder's edict: every player, the Devil's controller included, is
// asked for a nontoken creature of their own (the Devil itself counts, being
// on the battlefield by then), in turn order, and they are sacrificed together.
// A player with none sacrifices nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b73bbdcc-e872-4da5-8d80-0fc9e52418bf",
		Name:         "Chain Devil",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.WhenThisEnters("Chain Devil — each player sacrifices a nontoken creature of their choice", effects.Do(effects.EachPlayerSacrifices{
				Match: effects.B04NontokenCreature,
				Label: "a nontoken creature",
			}))),
		},
	})
}
