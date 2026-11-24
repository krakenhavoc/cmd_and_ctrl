package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Creative Outburst — Instant {3}{U}{U}{R}{R} (Strixhaven: School of Mages,
// #5502):
//
//	"Creative Outburst deals 5 damage to any target. Look at the top five
//	 cards of your library. Put one of them into your hand and the rest on the
//	 bottom of your library in a random order.
//	 {U/R}{U/R}, Discard this card: Create a Treasure token."
//
// The spell is Prophetic Titan's second mode after a fixed 5 damage: a look
// (not a reveal) at the top five, one card mandatory, the rest on the bottom
// in a random order. The second ability is a hand-zone activated ability like
// Channel and cycling ({U/R}{U/R} is two hybrid payments, discard this card as
// a cost, CR 602.2); it creates one Treasure and does not cast the card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ffc79d71-5a1f-40a8-80af-84c85e12ba6e",
		Name:         "Creative Outburst",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DamageToFirstTargetStep(5)(item, ctx); err != nil {
				return err
			}
			player := item.Controller
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 5),
				Max:    1,
				Label:  "Creative Outburst — put one into your hand",
				Key:    effects.TakeRestBottomRandom,
			}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U/R}{U/R}, Discard this card: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{U/R}{U/R}"), effects.DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
