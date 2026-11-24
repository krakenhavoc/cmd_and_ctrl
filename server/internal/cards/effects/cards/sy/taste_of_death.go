package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Taste of Death — Sorcery {4}{B}{B} (Commander 2021, #5504):
//
//	"Each player sacrifices three creatures of their choice. You create
//	 three Food tokens."
//
// Every player still in the game is asked for three creatures, in APNAP
// order and each seeing the earlier choices (Myrkul's Edict's ordered
// collector); a player with fewer than three sacrifices what they have.
// The Foods come after the sacrifices, as printed, and are made whether or
// not anyone sacrificed anything.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f92673fe-cd73-48ae-bfbd-aa750582b80b",
		Name:         "Taste of Death",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var players []uuid.UUID
			for _, p := range ctx.Game.Seats {
				if p == nil || p.Eliminated {
					continue
				}
				players = append(players, p.ID, p.ID, p.ID)
			}
			return ctx.PlayersSacrificeThen(players, effects.SacrificeSpec("a creature", effects.Creature()),
				"Taste of Death — sacrifice three creatures", tasteOfDeathSacrificed, ctx.RunData(nil))
		},
	})
}

// tasteOfDeathSacrificed is the second sentence: you create three Foods.
var tasteOfDeathSacrificed = effects.RunThen("prompt-run/taste-of-death-sacrificed", func(ctx *effects.Context, _ game.RunDone) error {
	return effects.CreateToken{Template: effects.FoodToken(), N: 3}.Apply(ctx)
})
