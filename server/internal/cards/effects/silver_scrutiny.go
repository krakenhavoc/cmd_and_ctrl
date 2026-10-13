package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silver Scrutiny — Sorcery {X}{U}{U}:
//
//	"You may cast this spell as though it had flash if X is 3 or less.
//	 Draw X cards."
//
// CR 601.3c: the caster begins the cast at instant speed and the
// announced X must then be 3 or less. That is an alternative cost that
// opens the flash window (AlternativeCost.AsThoughFlash) carrying a
// ceiling on the X announced under it (AlternativeCost.XCeiling, #3225),
// so claiming the offer at instant speed refuses X = 4 while the plain
// sorcery-speed cast for the printed cost takes any X. The offer charges
// the printed cost; see FlashIfXAtMost.
func init() {
	Register(Spec{
		OracleID:         "2b0b4fc1-9ad6-4259-a0a2-9a14484c246e",
		Name:             "Silver Scrutiny",
		XMatters:         true,
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{FlashIfXAtMost("{X}{U}{U}", 3)},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: item.Controller, N: ctx.X()}.Apply(ctx)
		},
	})
}
