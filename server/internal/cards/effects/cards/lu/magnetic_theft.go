package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magnetic Theft — Instant {R}:
//
//	"Attach target Equipment to target creature. (Control of the
//	 Equipment doesn't change.)"
//
// Neither slot says "you control" — this moves anyone's sword onto
// anyone's creature — and the reminder text is not a separate clause
// to implement: AttachForEffect (game/attach.go) only ever writes the
// AttachedTo link, never Controller, so leaving both predicate lists
// empty already produces exactly what the parenthetical promises.
// Shares its two-slot clause shape and its resolution with Brass
// Squire's activated ability (attachments.go, ADR 0065).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b22cc97-003f-4227-acdc-7a0857674b67",
		Name:         "Magnetic Theft",
		Completeness: effects.CompletenessFull,
		Targets: effects.TwoSlotAttachTargets(
			"target Equipment", nil,
			"target creature", nil,
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.AttachClauseTargets(ctx.Game, item)
		},
	})
}
