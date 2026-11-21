package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Focus the Mind — Instant {4}{U} (Tarkir: Dragonstorm, #4463):
//
//	"This spell costs {2} less to cast if you've cast another spell this
//	 turn.
//	 Draw three cards, then discard a card."
//
// The reduction reads the caster's cast tally as the cost is determined
// (CR 601.2f). The spell being cast is not counted yet, because a spell
// becomes cast only at CR 601.2i, so any spell in the tally is another
// one. The body is Ancestral Reminiscence's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "695a84b5-96b0-4e48-b615-15624efab3b7",
		Name:         "Focus the Mind",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3, Discards: 1},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast if you've cast another spell this turn.",
				effects.TdmCastAnotherSpellThisTurn),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 3, 1, "")
		},
	})
}
