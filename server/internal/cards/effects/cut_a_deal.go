package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cut a Deal — Sorcery, {2}{W} (EDHREC rank 996):
//
//	"Each opponent draws a card, then you draw a card for each
//	 opponent who drew a card this way."
//
// White's group-hug draw three: everyone gets a card, and at a full
// table the caster gets three. The opponents draw in seat order,
// each an ordinary draw (Sheoldred and Razorkin Needlehead see every
// one), and the caster draws the count of opponents who actually
// had a card to draw — an opponent with an empty library draws
// nothing and earns nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b5cffd5-5db3-4151-9275-d91387554412",
		Name:         "Cut a Deal",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			drew, err := b08EachOpponentDraws(ctx.Game, item)
			if err != nil {
				return err
			}
			return DrawCards{Player: item.Controller, N: drew}.Apply(ctx)
		},
	})
}
