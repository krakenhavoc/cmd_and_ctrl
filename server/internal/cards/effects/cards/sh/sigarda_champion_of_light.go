package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sigarda, Champion of Light — Legendary Creature — Angel {1}{G}{W}{W}, 4/4
// (Midnight Hunt Commander, S82 MID batch 13):
//
//	"Flying, trample
//	 Humans you control get +1/+1.
//	 Coven — Whenever Sigarda attacks, if you control three or more creatures
//	 with different powers, look at the top five cards of your library. You
//	 may reveal a Human creature card from among them and put it into your
//	 hand. Put the rest on the bottom of your library in a random order."
//
// The anthem is a layer 7c modify over the Humans the controller controls
// ("Human" is a word of the source's text, CR 612.2). The attack trigger is
// Redemption Choir's coven shape, an intervening "if" (CR 603.4) on three
// distinct powers through the layers, checked as Sigarda attacks and again as
// the trigger resolves, with Frontier Seeker's whole look-at sentence
// (LookAtTopThenMayTakeToHand).
//
// No simplification.
func init() {
	coven := effects.CondBoard{
		Name: "you control three or more creatures with different powers",
		Test: func(ctx *effects.Context) bool { return effects.TrekB48Coven(ctx.Game, ctx.Controller(), ctx.Source()) },
	}
	effects.Register(effects.Spec{
		OracleID:        "24919141-82d2-42ad-963c-e53511a868b3",
		Name:            "Sigarda, Champion of Light",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			Label:    "Humans you control get +1/+1.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Human"))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(coven, effects.On(game.EventAttack,
				func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
					return effects.ThisAttacked(ev, source, ch, g) && effects.TrekB48Coven(g, source.Controller, source.InstanceID)
				},
				"Sigarda, Champion of Light — Coven: look at the top five cards of your library; you may reveal a Human creature card from among them and put it into your hand; the rest on the bottom in a random order",
				effects.LookAtTopThenMayTakeToHand(5, effects.And(effects.Creature(), effects.HasSubtype("Human")), 1,
					"Sigarda, Champion of Light — you may reveal a Human creature card from among them and put it into your hand"))),
		},
	})
}
