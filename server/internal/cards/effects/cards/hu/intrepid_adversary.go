package cardshu

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Intrepid Adversary — Creature — Human Scout {1}{W}, 3/1:
//
//	"Lifelink
//	 When this creature enters, you may pay {1}{W} any number of times.
//	 When you pay this cost one or more times, put that many valor
//	 counters on this creature.
//	 Creatures you control get +1/+1 for each valor counter on this
//	 creature."
//
// The Adversary cycle's shape (Bloodthirsty Adversary) with a custom
// "valor" counter, so Hardened Scales does nothing to it and proliferate
// grows it. The anthem is a layer 7c static read live off the counters
// (Boon of the Spirit Realm); it includes the Adversary itself.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b72fde5e-862a-45e8-9880-f2c16975a9ac",
		Name:            "Intrepid Adversary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Intrepid Adversary — pay {1}{W} any number of times; that many valor counters", intrepidAdversaryEnters),
		},
		Static: []game.StaticAbility{{
			Label:    "Creatures you control get +1/+1 for each valor counter on this creature.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					source.Counters["valor"] > 0
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				n := source.Counters["valor"]
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}

var intrepidAdversaryBody = game.ReflexiveBody("intrepid-adversary/valor",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		return effects.CounterOnThis{Kind: "valor", N: p.Amount}.Apply(effects.NewContext(g, item))
	}, nil)

var intrepidAdversaryPaid = effects.PayUnlessBranches("pay-unless/intrepid-adversary-times",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return effects.ReflexiveTrigger{
			Label:  "Intrepid Adversary — " + strconv.Itoa(r.Number) + " valor counters",
			Body:   intrepidAdversaryBody,
			Params: game.EffectParams{Amount: r.Number},
		}.Apply(ctx)
	}, nil)

func intrepidAdversaryEnters(g *game.Game, item *game.StackItem) error {
	return effects.MayPayX{
		Symbol: "{1}{W}",
		Label:  "Intrepid Adversary",
		Buys:   "put X valor counters on it",
		Goal:   effects.XAsHighAsYouCan,
		Then:   intrepidAdversaryPaid,
	}.Apply(effects.NewContext(g, item))
}
