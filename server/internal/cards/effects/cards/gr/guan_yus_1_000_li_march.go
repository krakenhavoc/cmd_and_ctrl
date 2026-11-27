package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guan Yu's 1,000-Li March — Sorcery {4}{W}{W} (Portal Three Kingdoms, S82 OLD b713, #5622):
//
//	"Destroy all tapped creatures."
//
// Righteous Fury's sweep without the life: every tapped creature of every
// player, yours included, is destroyed at once; tapped is read as the spell
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1b87d5d-9493-4001-8c7e-553f40de61ea",
		Name:         "Guan Yu's 1,000-Li March",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.TappedPermanent())}.Apply(ctx)
		},
	})
}
