package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shrine of Piercing Vision — Artifact {2} (New Phyrexia):
//
//	"At the beginning of your upkeep and whenever you cast a blue spell,
//	 put a charge counter on this artifact.
//	 {T}, Sacrifice this artifact: Look at the top X cards of your library, where X is the number of charge counters on this artifact. Put one of those cards into your hand and the rest on the bottom of your library in any order."
//
// The counter clause is ONE triggered ability with two trigger conditions.
// The second ability sacrifices the Shrine as a cost, so the number of
// charge counters is its last-known information (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9017da9a-2cd5-42fe-a82d-63c34e952344",
		Name:         "Shrine of Piercing Vision",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB460ShrineCounterTrigger("Shrine of Piercing Vision", "U"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this artifact: Look at the top X cards of your library, where X is the number of charge counters on this artifact. Put one of those cards into your hand and the rest on the bottom of your library in any order.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := effects.OldB460ChargeCountersOnSource(g, item.SourceCardID)
				if x <= 0 {
					return nil
				}
				player := ctx.Controller()
				return effects.TakeFromLibraryToHand{
					Player: player,
					Cards:  g.LookAtTopOfLibraryForEffect(player, x),
					Max:    1,
					Label:  "Shrine of Piercing Vision — put one of those cards into your hand",
					Key:    effects.TakeRestBottomAny,
				}.Apply(ctx)
			},
		}},
	})
}
