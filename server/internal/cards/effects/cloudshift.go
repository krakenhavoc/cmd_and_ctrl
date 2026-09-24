package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloudshift — Instant {W}:
//
//	"Exile target creature you control, then return that card to the
//	 battlefield under your control."
//
// "Under YOUR control" rather than "under its owner's control" is
// the whole reason Flicker.Controller exists: the target and the
// destination controller are usually the same seat, but not when the
// creature is under a control-magic effect the caster doesn't own —
// Cloudshift ends the theft (CR 400.7 makes it a new object, entering
// under the CASTER'S control) where Essence Flux (owner's control)
// would hand it back.
//
// The permanent returns as a NEW OBJECT — fresh InstanceID, no
// counters, no damage, summoning-sick again — and every ETB it has
// re-triggers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6879f5ce-7a1b-4606-bad1-885779b0d456",
		Name:         "Cloudshift",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			return Flicker{Target: item.Targets[0].ID, Controller: ctx.Controller()}.Apply(ctx)
		},
	})
}
