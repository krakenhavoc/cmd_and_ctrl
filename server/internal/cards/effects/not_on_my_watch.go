package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Not on My Watch — Instant {1}{W} (Murders at Karlov Manor, #4469):
//
//	"Exile target attacking creature."
//
// Gaze of Justice's exile with the target narrowed to an attacking
// creature. A creature removed from combat in response is no longer a
// legal target, so the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2d400c01-d2d0-442a-bc3d-cf106a754dab",
		Name:         "Not on My Watch",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking creature", AttackingCreature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
