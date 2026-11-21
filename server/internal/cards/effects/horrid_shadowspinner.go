package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Horrid Shadowspinner — Creature — Horror {1}{U}{B}, 2/3 (Modern Horizons 3):
//
//	"Lifelink
//	 Whenever this creature attacks, you may draw cards equal to its power. If
//	 you do, discard that many cards."
//
// Its power is read as the ability resolves, from last-known information if it
// has left (CR 608.2h); the discard is the same number.
func init() {
	Register(Spec{
		OracleID:        "5fcbb5d3-71d1-4946-b806-ba6376338eba",
		Name:            "Horrid Shadowspinner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Horrid Shadowspinner — you may draw cards equal to its power; if you do, discard that many cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					power := 0
					if ok && info.Power > 0 {
						power = info.Power
					}
					return MayChoice{
						Question: "Horrid Shadowspinner — draw " + strconv.Itoa(power) + " cards, then discard that many cards?",
						Draws:    power,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     horridShadowspinnerAnswer,
						Word:     strconv.Itoa(power),
					}.Apply(ctx)
				}),
		},
	})
}

var horridShadowspinnerAnswer = MayYes("confirm/horrid-shadowspinner-loot", func(ctx *Context, r game.ConfirmAnswer) error {
	n := wordInt(r)
	return drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), n, n, "")
})
