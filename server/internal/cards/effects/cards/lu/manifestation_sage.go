package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Manifestation Sage — Creature — Human Wizard {G/U}{G/U}{G/U}{G/U}, 2/2
// (Strixhaven: School of Mages, #5502):
//
//	"When this creature enters, create a 0/0 green and blue Fractal creature
//	 token. Put X +1/+1 counters on it, where X is the number of cards in your
//	 hand."
//
// Deekah's create-then-put shape (SocB8FractalWithCounters): the Fractal is made
// first and the counters are then put on it, X counted once as the ability
// resolves (CR 608.2h). With an empty hand the Fractal is a 0/0 and dies to
// CR 704.5f.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8bcd896-cae9-4733-a437-165644b8b0e8",
		Name:         "Manifestation Sage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Manifestation Sage — create a 0/0 Fractal and put X +1/+1 counters on it, X the cards in your hand",
				effects.SocB8FractalWithCounters("manifestation-sage/fractal", func(ctx *effects.Context) int {
					return effects.CardsInHand(ctx.Game, ctx.Controller())
				})),
		},
	})
}
