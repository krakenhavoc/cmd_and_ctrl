package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disrupt Decorum — Sorcery, {2}{R}{R}:
//
//	"Goad all creatures you don't control. (Until your next turn,
//	those creatures attack each combat if able and attack a player
//	other than you if able.)"
//
// #1599: GoadAllMatching (goad.go) over every creature the caster
// doesn't control, evaluated at resolution — the affected set is
// whoever is out there right now, per CR 701.15's own "goad" (a
// one-shot marker, not a locked continuous effect). One delayed
// trigger clears every stamped creature's marker at the beginning of
// the caster's next turn, same as Alela's per-attack goad.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9b084cd4-cc21-4f3a-9ac5-1e1160a31aa6",
		Name:         "Disrupt Decorum",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GoadAllMatching(ctx, func(_ *game.Game, c game.Card) bool {
				return c.Controller != ctx.Controller()
			})
		},
	})
}
