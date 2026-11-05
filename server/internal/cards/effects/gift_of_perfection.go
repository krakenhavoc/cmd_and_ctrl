package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gift of Perfection — Sorcery {4}{U} (Star Trek Commander, #4974):
//
//	"Each opponent reveals cards from the top of their library until they
//	 reveal a creature card. Assimilate one of the creature cards revealed
//	 this way. Then each opponent puts the other cards they revealed this
//	 way on the bottom of their library in a random order. (To assimilate
//	 a card, put it onto the battlefield under your control with a +1/+1
//	 counter. It's a Borg artifact creature and loses all other creature
//	 types.)"
//
// trekB34GiftOfPerfection: every opponent reveals, the caster picks one of
// the creature cards, it is assimilated out of its owner's library, and
// every other revealed card goes to the bottom of its owner's library in a
// random order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b44026ae-eaf1-4daf-99c6-2663db674ff7",
		Name:         "Gift of Perfection",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return trekB34GiftOfPerfection(ctx.Game, item)
		},
	})
}
