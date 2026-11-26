package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fated Infatuation — Instant {U}{U}{U} (Born of the Gods, S82 OLD b377, #5559):
//
//	"Create a token that's a copy of target creature you control. If it's your
//	 turn, scry 2."
//
// Quasiduplicate's token copy, then the scry behind "if it's your turn", read
// as the spell resolves (Fated Conflagration's check).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae771e78-1d29-4b8c-8511-9da196a78dcf",
		Name:         "Fated Infatuation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.TokenCopyOfSingleTarget(item, ctx); err != nil {
				return err
			}
			if !effects.IsYourTurn(ctx.Game, item.Controller) {
				return nil
			}
			return effects.Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
