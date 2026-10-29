package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Uncover the Moon-Letters — Enchantment {3}{U} (The Hobbit):
//
//	"Whenever you cast a noncreature spell, you may draw X cards, where X is
//	 the amount of mana spent to cast that spell. If you do, discard two
//	 cards."
//
// Aetherflux Conduit's trigger: X is read off the payment record as the spell
// is cast (ReadsManaSpent, so a waived cast asks what was paid outside the
// pool) and carried on the item, so a spell countered in response still pays
// out. A spell cast for no mana is not given a trigger: drawing no cards is
// not "doing" it, so there would be nothing to discard for. The "you may" is
// asked as the ability resolves (CR 603.5), and the draw and the two discards
// are one linked clause (The Celestus' shape): declining the draw declines
// the discard, and a player a draw limit stops from drawing X cards is not
// asked (MayChoice.Draws). The discard opens after every draw has landed, so
// a drawn card is a legal discard, and a hand of fewer than two discards what
// it has (CR 701.9b).
//
// No simplification.
func init() {
	const label = "Uncover the Moon-Letters — you may draw X cards, where X is the amount of mana spent to cast that spell; if you do, discard two cards"
	t := On(game.EventCast, AllOf(YouCast(Noncreature()), uncoverCastSpentMana), label,
		func(g *game.Game, item *game.StackItem) error {
			x := item.Params.Amount
			if x <= 0 {
				return nil
			}
			return MayChoice{
				Question: "Uncover the Moon-Letters — draw " + strconv.Itoa(x) + " cards, then discard two cards?",
				Draws:    x,
				YesLabel: "Draw",
				NoLabel:  "Decline",
				Word:     strconv.Itoa(x),
				Then:     uncoverDrawThenDiscard,
			}.Apply(NewContext(g, item))
		})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	Register(Spec{
		OracleID:       "4c67359d-5590-4eb0-827e-b411967ee238",
		Name:           "Uncover the Moon-Letters",
		ReadsManaSpent: game.ReadsYourCasts,
		Completeness:   CompletenessFull,
		Triggered:      []game.TriggeredAbility{t},
	})
}

// uncoverCastSpentMana keeps the trigger off the stack for a spell cast for no
// mana: there are no cards to draw.
func uncoverCastSpentMana(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	return g.StackItemPaidForEffect(ev.CardID).Spent().Total() > 0
}

// uncoverDrawThenDiscard is the "yes": draw X cards, then discard two. Keyed
// (#4291); the amount rides in the answer's word.
var uncoverDrawThenDiscard = MayYes("confirm/uncover-the-moon-letters-draw", func(ctx *Context, r game.ConfirmAnswer) error {
	return drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), wordInt(r), 2, "")
})
