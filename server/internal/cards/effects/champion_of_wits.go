package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Champion of Wits — Creature — Snake Wizard {2}{U}, 2/1:
//
//	"When this creature enters, you may draw cards equal to its power. If
//	 you do, discard two cards.
//	 Eternalize {5}{U}{U}"
//
// One linked "you may … if you do" clause (CR 607.2) asked as the trigger
// resolves. The power is read then, from the creature as it last existed if
// it has left (CR 608.2h), and handed to the keyed answer as its Word. The
// discard is two cards whatever the draw was, and it opens only after the
// draw has resolved. An eternalized token enters as a 4/4, so it draws four.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3fd8ba0c-75ec-43ea-8706-f71e6cf1d236",
		Name:         "Champion of Wits",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Champion of Wits — you may draw cards equal to its power; if you do, discard two cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					power := 0
					if ok && info.Power > 0 {
						power = info.Power
					}
					return MayChoice{
						Question: "Champion of Wits — draw " + strconv.Itoa(power) + " cards, then discard two cards?",
						Draws:    power,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     championOfWitsAnswer,
						Word:     strconv.Itoa(power),
					}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{Eternalize("{5}{U}{U}")},
	})
}

// championOfWitsAnswer is the "yes": draw N cards (N is the Word), then
// discard two. Keyed.
var championOfWitsAnswer = MayYes("confirm/champion-of-wits-loot", func(ctx *Context, r game.ConfirmAnswer) error {
	return drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), wordInt(r), 2, "")
})
