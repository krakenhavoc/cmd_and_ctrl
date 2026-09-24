package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Void Rend — Instant {W}{U}{B}:
//
//	"This spell can't be countered.
//	 Destroy target nonland permanent."
//
// Abrupt Decay without the mana-value cap, at triple the color
// commitment. Same two pieces: Spec.CantBeCountered (S23) and a
// Nonland() target clause.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "713f16db-95ec-479e-a48c-7a69f7668d7f",
		Name:            "Void Rend",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		Targets:         TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyChosenPermanent(ctx.Game, item)
		},
	})
}
