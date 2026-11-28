package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tariff — Sorcery {1}{W} (Weatherlight, S82 older sets batch 779, #5629):
//
//	"Each player sacrifices the creature they control with the greatest mana value unless they pay
//	 that creature's mana cost. If two or more creatures a player controls are tied for greatest, that
//	 player chooses one."
//
// In turn order from the active player each player picks among their creatures tied for greatest mana value (a player with one is not asked)
// and is asked to pay its mana cost; declining, or being unable to pay, sacrifices it. A creature with no mana cost (a token, say)
// has nothing to pay, so it is sacrificed (CR 118.3). Players with no creature are unaffected.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e96b0560-1d91-48df-b42f-3b7d9c6d94bd",
		Name:         "Tariff",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB779Tariff(ctx)
		},
	})
}
