package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Powerstone Shard — Artifact {3} (Dominaria, S82 OLD, #5530):
//
//	"{T}: Add {C} for each artifact you control named Powerstone Shard."
//
// A count-driven mana ability (Cloudpost's ProducedAmount) read as it is
// activated: every artifact the Shard's controller controls with that name,
// itself included, so one Shard makes {C} and two make {C}{C} each.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "769c45ae-5d49-4815-944e-98790ad4efb0",
		Name:         "Powerstone Shard",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:  effects.ManaAbilityCost{Tap: true},
			Label: "{T}: Add {C} for each artifact you control named Powerstone Shard",
			ProducedAmount: func(g *game.Game, controller, _ uuid.UUID) game.ManaAmount {
				return game.ManaAmount{Options: []string{"C"}, N: effects.S82OldB213PowerstoneShards(g, controller)}
			},
		}},
	})
}
