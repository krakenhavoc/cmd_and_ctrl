package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reject — Instant {1}{U} (Strixhaven: School of Mages, #5502):
//
//	"Counter target creature or planeswalker spell unless its controller pays
//	 {3}. If that spell is countered this way, exile it instead of putting it
//	 into its owner's graveyard."
//
// Spell Pierce's pay-unless prompt on a creature or planeswalker spell, with
// Syncopate's exile destination: the destination applies only when the spell
// really is countered, so a spell whose controller pays {3} resolves and one
// that can't be countered is left alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "198fed4c-297e-41c0-a172-e471c7401fbb",
		Name:         "Reject",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target creature or planeswalker spell", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			return effects.CounterUnlessPaid{
				StackID:  targets[0].ID,
				Cost:     "{3}",
				Question: "Reject — pay {3} or your spell is countered and exiled",
				Dest:     game.ZoneRef{Kind: game.ZoneExile},
			}.Apply(ctx)
		},
	})
}
