package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Redirect — Instant {U}{U}:
//
//	"You may choose new targets for target spell."
//
// CR 115.7c with the "you may" of Deflecting Swat: every target slot of
// the spell may move, and the controller may decline. Unlike
// Deflecting Swat the clause has no "or ability" half and no
// single-target restriction — a two-target spell is a legal target and
// both of its targets may be rechosen. Nothing is simplified.
func init() {
	Register(Spec{
		OracleID:     "74a7d880-5a49-412b-af4d-0b4cdb8453ee",
		Name:         "Redirect",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return chooseNewTargets("Redirect — choose new targets")(ctx.Game, item)
		},
	})
}
