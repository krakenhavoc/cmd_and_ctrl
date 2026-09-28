package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Falter — Instant {1}{R}:
//
//	"Creatures without flying can't block this turn."
//
// The proof card for #1650 (ADR 0045's amendment of 2026-09-28). The
// affected set is a RULE, not a list: "can't block" changes no
// characteristic, so CR 611.2c does not lock it to the creatures on the
// battlefield as Falter resolves. A creature flashed in afterwards, or
// one that loses flying, can't block either; one that gains flying
// after Falter resolved can. RestrictUntilEOT's Scope form reads the
// set live until cleanup, against each creature's finished
// characteristics.
//
// Magmatic Chasm and Seismic Stomp print the same line at sorcery
// speed and share faltering().
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d4b50749-a016-4aff-8d70-a1707cabf57b",
		Name:         "Falter",
		Completeness: CompletenessFull,
		OnResolve:    faltering("Falter"),
	})
}

// faltering is "Creatures without flying can't block this turn." as a
// spell's whole resolution — Falter, Magmatic Chasm, Seismic Stomp.
func faltering(name string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		return RestrictUntilEOT{
			Scope:        game.ScopeCreaturesWithoutFlying,
			Restrictions: game.CantBlock,
			Label:        name + " — creatures without flying can't block",
		}.Apply(ctx)
	}
}
