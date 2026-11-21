package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spatial Contortion — Instant {1}{C} (Commander Masters, #5948):
//
//	"Target creature gets +3/-3 until end of turn."
//
// The {C} in the cost is colourless mana only (CR 107.4c); the effect is a
// layer 7c modification until end of turn (CR 613.4c), so a creature whose
// toughness reaches zero dies as state-based actions are checked.
//
// No Purpose: the change is partly a benefit and partly a harm, which
// TargetPurpose.UntilEndOfTurn leaves undeclared (#4312).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7346337e-ac12-4dd2-826d-77ec0013e74f",
		Name:         "Spatial Contortion",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return boostTargetUntilEOT("Spatial Contortion — +3/-3", 3, -3)(ctx.Game, ctx.Item)
		},
	})
}
