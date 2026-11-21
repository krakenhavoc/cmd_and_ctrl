package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirrodin Avenged — Instant {B} (March of the Machine, S82 #5477):
//
//	"Destroy target creature that was dealt damage this turn.
//	 Draw a card."
//
// The target restriction is read both when the spell is cast and when it
// resolves (CR 608.2b). If the creature is no longer a legal target the
// spell does not resolve, so no card is drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec83f398-31af-484d-b6d2-8c7add142260",
		Name:         "Mirrodin Avenged",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetCreature("target creature that was dealt damage this turn", effects.DealtDamageThisTurn()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				break
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
