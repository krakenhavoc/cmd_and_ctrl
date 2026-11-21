package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Telekinesis — Instant {U}{U}:
//
//	"Tap target creature. Prevent all combat damage that would be dealt by
//	 that creature this turn. It doesn't untap during its controller's next
//	 two untap steps."
//
// #2029 (ADR 0058, 2026-10-08 amendment): the marker counts steps. It is
// keyed to "its controller", not to the player who cast the spell, so
// the effect does not lock in whoever controls the creature now: if
// control changes it still skips the next two untap steps of whoever
// controls it at each one (the ruling). The shield half is ADR 0108's
// source shield, pinned to the targeted creature as the spell resolves.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b9dd2b6-d14d-4c1e-9885-00ab2c0bf8da",
		Name:         "Telekinesis",
		Purpose:      game.Purpose{Answers: game.AnswerPrevent}, // #4734 slice 6
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 || targets[0].Kind != game.TargetCard {
				return nil
			}
			id := targets[0].ID
			if err := ctx.Game.TapTargetForEffect(id); err != nil {
				return err
			}
			if err := (effects.PreventDamageFromSource{From: id, CombatOnly: true, Protect: effects.ShieldAnything}).Apply(ctx); err != nil {
				return err
			}
			return effects.DoesntUntapNextUntapStep{Targets: []uuid.UUID{id}, Steps: 2, Label: "Telekinesis"}.Apply(ctx)
		},
	})
}
