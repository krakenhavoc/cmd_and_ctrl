package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yawgmoth's Will — Sorcery {2}{B}:
//
//	"Until end of turn, you may play lands and cast spells from your
//	 graveyard.
//	 If a card would be put into your graveyard from anywhere this turn,
//	 exile that card instead."
//
// Both halves are one GraveyardPlayThisTurn, so the permission cannot
// ship without the replacement (ADR 0108 §4). The Will exiles itself:
// the replacement is registered while it resolves, and it goes to the
// graveyard after.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "322f0459-f394-44f0-977b-55fd0cbe0712",
		Name:         "Yawgmoth's Will",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.YawgmothsWillThisTurn().Apply(ctx)
		},
	})
}
