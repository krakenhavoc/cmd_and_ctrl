package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Book Devourer — Creature — Beast {5}{R}, 4/5 (Guilds of Ravnica):
//
//	"Trample
//	 Whenever this creature deals combat damage to a player, you may discard
//	 all the cards in your hand. If you do, draw that many cards."
//
// Phyrexian Dragon Engine's "you may discard your hand" answer, drawing the
// number of cards actually discarded (DiscardWholeHand's count) instead of a
// fixed three. An empty hand may still be "discarded" and draws nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fb6d24cb-4e79-4d0c-b32f-47d5f709de1b",
		Name:            "Book Devourer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Book Devourer — you may discard all the cards in your hand; if you do, draw that many cards",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Book Devourer — discard all the cards in your hand? If you do, draw that many cards.",
						YesLabel: "Discard your hand and draw that many",
						NoLabel:  "Keep your hand",
						Then:     bookDevourerDiscard,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// bookDevourerDiscard is the yes: discard the whole hand, draw that many. An
// on-disk key.
var bookDevourerDiscard = effects.MayYes("confirm/book-devourer-discard-hand", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	n, err := effects.DiscardWholeHand(ctx.Game, r.Chooser)
	if err != nil {
		return err
	}
	return effects.DrawCards{Player: r.Chooser, N: n}.Apply(ctx)
})
