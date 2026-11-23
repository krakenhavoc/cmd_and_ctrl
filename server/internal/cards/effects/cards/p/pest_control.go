package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pest Control — Sorcery {W}{B}:
//
//	"Destroy all nonland permanents with mana value 1 or less.
//	 Cycling {2}"
//
// Mana value is read off the permanent as it is (CR 202.3): {X} counts
// as 0 and a token that copies nothing has mana value 0, so both go
// (the rulings).
//
// No simplification. Star Trek Commander (#4974).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "072db429-3838-42ba-89b6-e26cbce22af9",
		Name:         "Pest Control",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{Sweep: game.Sweep{
			Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true,
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Nonland(), effects.ManaValueLE(1))}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
