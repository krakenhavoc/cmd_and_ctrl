package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sealed Fate — Sorcery {X}{U}{B} (Mirage, S82 OLD b798, #5633):
//
//	"Look at the top X cards of target opponent's library. Exile one of those
//	 cards and put the rest back on top of that player's library in any
//	 order."
//
// Gonti's look at an opponent's library: only the caster sees the cards (CR
// 701.20), picks the one to exile (face up) and orders the rest back on top of
// that player's library. A library of one card exiles it with no choice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b9b5856-9396-4a55-8675-58cd171df708",
		Name:         "Sealed Fate",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		Purpose:      game.Purpose{Answers: game.AnswerRestrict},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.OldB798SealedFate(ctx.Game, item)
		},
	})
}
