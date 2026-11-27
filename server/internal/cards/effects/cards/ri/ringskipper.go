package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ringskipper — Creature — Faerie Wizard {1}{U}, 1/1 (Lorwyn, #5593):
//
//	"Flying
//	 When this creature dies, clash with an opponent. If you win, return
//	 this card to its owner's hand. (Each clashing player reveals the top
//	 card of their library, then puts that card on their choice of the top
//	 or bottom. A player wins if their card had a greater mana value.)"
//
// Flying is a printed keyword. The dies trigger clashes, and the return is
// the clash's tail: it finds the card in the graveyard only if it is still
// the same object that died (CR 400.7), so a card that left and came back
// is not returned.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8e353b5f-0692-430e-9558-59818275b13b",
		Name:            "Ringskipper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Ringskipper — clash with an opponent; if you win, return this card to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ClashWithAnOpponent{Then: ctx.ItemTail(ringskipperClashed)}.Apply(ctx)
				}),
		},
	})
}

// ringskipperClashed is "If you win, return this card to its owner's hand".
// An on-disk key.
var ringskipperClashed = effects.CardTail("then/ringskipper-clashed", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.WonTheClash(o) {
		return nil
	}
	return effects.ReturnSelfFromGraveyardToHand(ctx.Game, ctx.Item)
})
