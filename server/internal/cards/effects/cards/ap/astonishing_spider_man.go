package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Astonishing Spider-Man — {2}{U/R}{U/R} Legendary Creature — Spider
// Human Hero 4/4 (Marvel Super Heroes Commander, #4455):
//
//	"Reach
//	 Whenever Astonishing Spider-Man attacks, you may discard your hand.
//	 If you do, draw a card for each card you've discarded this turn."
//
// A keyed yes/no (Academic Dispute's MayYes) on the attack trigger. On a
// yes the whole hand is discarded (discardWholeHand), and the draw reads
// the turn's discard tally as Misty Knight's does, so the hand just
// discarded counts with every card discarded earlier in the turn. A
// player with no hand discards nothing, so "if you do" is not met and
// nothing is drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "24059283-edbd-4817-9f35-4be02f07b8f8",
		Name:            "Astonishing Spider-Man",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Astonishing Spider-Man — you may discard your hand; if you do, draw a card for each card you've discarded this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Astonishing Spider-Man — discard your hand? If you do, draw a card for each card you've discarded this turn.",
						Then:     astonishingSpiderManDiscardHand,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var astonishingSpiderManDiscardHand = effects.MayYes("confirm/astonishing-spider-man-discard-hand", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	n, err := effects.DiscardWholeHand(ctx.Game, r.Chooser)
	if err != nil || n == 0 {
		return err
	}
	return effects.DrawCards{Player: r.Chooser, N: ctx.Game.TurnTallyFor(r.Chooser).CardsDiscarded}.Apply(ctx)
})
