package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Merchant's Dockhand — Artifact Creature — Construct {1}, 1/2 (Aether
// Revolt, S82 older sets, #5536):
//
//	"{3}{U}, {T}, Tap X untapped artifacts you control: Look at the top X
//	 cards of your library. Put one of them into your hand and the rest on
//	 the bottom of your library in any order."
//
// X is the number of artifacts tapped to pay (CR 601.2b). The Dockhand's own
// {T} is a separate cost, so it is not one of the X (an artifact creature
// tapped for {T} is already tapped and cannot be tapped again, CR 118.3).
// Taking one is mandatory when there is a card to take; the rest go to the
// bottom in the order the player chooses (CR 701.20). With X of 0 nothing is
// looked at.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7cd95b92-8828-4b84-be58-4ce3e948dc8c",
		Name:         "Merchant's Dockhand",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}, {T}, Tap X untapped artifacts you control: Look at the top X cards of your library. Put one of them into your hand and the rest on the bottom of your library in any order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost: effects.Plus(effects.ManaCost("{3}{U}"), effects.TapCost(),
				effects.TapXUntapped("X untapped artifacts you control", effects.Artifact())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				return effects.TakeFromLibraryToHand{
					Player: ctx.Controller(),
					Cards:  g.LookAtTopOfLibraryForEffect(ctx.Controller(), x),
					Max:    1,
					Label:  "Merchant's Dockhand — put one of them into your hand",
					Key:    effects.TakeRestBottomAny,
				}.Apply(ctx)
			},
		}},
	})
}
