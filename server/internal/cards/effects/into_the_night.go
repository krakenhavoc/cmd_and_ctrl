package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Into the Night — {3}{R} Sorcery (#2561, ADR 0132):
//
//	"It becomes night. Discard any number of cards, then draw that many
//	 cards plus one."
//
// "It becomes night" comes first, so a Celestus or a Brimstone Vandal
// triggers off it above the discard. The draw is the count actually
// discarded plus one; with an empty hand there is nothing to discard and
// it just draws the one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "58c131f3-c623-4324-b061-9535018676b6",
		Name:         "Into the Night",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (BecomeNight{}).Apply(ctx); err != nil {
				return err
			}
			hand := len(allHandCardIDs(ctx.Game, item.Controller))
			if hand == 0 {
				return ctx.Game.DrawNForEffect(item.Controller, 1)
			}
			return b39MayDiscardThenDraw(hand, false,
				"Into the Night — discard any number of cards, then draw that many plus one",
				func(discarded int) int { return discarded + 1 })(ctx)
		},
	})
}
