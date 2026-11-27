package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sea God's Revenge — Sorcery {5}{U} (Theros, S82 old sets):
//
//	"Return up to three target creatures your opponents control to their
//	 owners' hands. Scry 1."
//
// Waterwhirl's simultaneous bounce over the targets still legal (CR 608.2b),
// then the scry, with the opponents-only restriction on the targets.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ab71a49-ce81-44f7-b940-50e7aba64df9",
		Name:         "Sea God's Revenge",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Targets:      effects.TargetCreature("up to three target creatures your opponents control", effects.OpponentControls()).WithCount(0, 3),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ctx.Game.BounceCardsToHandForEffect(effects.LegalTargetCardIDs(ctx))
			return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
