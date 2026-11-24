package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glacian, Powerstone Engineer — Legendary Creature — Human Artificer
// {5}{U}, 3/6 (Commander Legends):
//
//	"{T}, Tap X untapped artifacts you control: Look at the top X cards of
//	 your library. Put one of those cards into your hand and the rest into
//	 your graveyard.
//	 Partner"
//
// Belisarius Cawl's variable "Tap X untapped ... you control" cost (the
// announced X is the width of the payment; Glacian's own {T} is separate, so
// he can't tap himself for it even if he is an artifact) with Codecracker
// Hound's look: exactly one of the X cards goes to the hand and the rest into
// the graveyard, not a mill. Partner is a deck-construction rule
// (CR 702.124).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e96ce18e-b002-4b5b-9560-c1634320e164",
		Name:         "Glacian, Powerstone Engineer",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Tap X untapped artifacts you control: Look at the top X cards of your library. Put one of those cards into your hand and the rest into your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.TapXUntapped("X untapped artifacts you control", effects.Artifact())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				player := ctx.Controller()
				return effects.TakeFromLibraryToHand{
					Player: player,
					Cards:  g.LookAtTopOfLibraryForEffect(player, x),
					Max:    1,
					Label:  "Glacian, Powerstone Engineer — put one of them into your hand and the rest into your graveyard",
					Key:    effects.TakeRestGraveyard,
				}.Apply(ctx)
			},
		}},
	})
}
