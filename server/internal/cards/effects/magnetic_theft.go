package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9b22cc97-003f-4227-acdc-7a0857674b67",
		Name:         "Magnetic Theft",
		Completeness: CompletenessFull,
		Targets: TwoSlotAttachTargets(
			"target Equipment", nil,
			"target creature", nil,
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return AttachClauseTargets(ctx.Game, item)
		},
	})
}
