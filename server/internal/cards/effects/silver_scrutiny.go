package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silver Scrutiny — Sorcery {X}{U}{U}:
//
//	"You may cast this spell as though it had flash if X is 3 or less.
//	 Draw X cards."
//
// CR 601.3b: the X the caster will announce can make the permission
// apply, so they may begin the cast at instant speed, and the announced
// X must then be 3 or less. That is the card's own timing permission
// with a ceiling on X (FlashIfXAtMost, Spec.FlashIf, #5207): a cast that
// is open only because of it refuses X = 4, while a sorcery-speed cast
// (or one another flash permission opens) takes any X. It is not an
// alternative cost, so a flash cast may still claim one (CR 118.9a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2b0b4fc1-9ad6-4259-a0a2-9a14484c246e",
		Name:         "Silver Scrutiny",
		XMatters:     true,
		Completeness: CompletenessFull,
		FlashIf:      FlashIfXAtMost(3),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: item.Controller, N: ctx.X()}.Apply(ctx)
		},
	})
}
