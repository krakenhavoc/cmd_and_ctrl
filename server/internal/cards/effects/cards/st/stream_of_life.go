package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stream of Life — Sorcery {X}{G} (Eighth Edition, #5607):
//
//	"Target player gains X life."
//
// Natural Spring's targeted gain with the amount the X announced as the spell was
// cast. A target that is no longer legal as it resolves gains nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9eb2912d-2130-49f2-9529-b58fa5a97a15",
		Name:         "Stream of Life",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			p, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok || ctx.X() <= 0 {
				return nil
			}
			return effects.GainLife{Player: p, Amount: ctx.X()}.Apply(ctx)
		},
	})
}
