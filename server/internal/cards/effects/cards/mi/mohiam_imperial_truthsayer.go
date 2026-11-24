package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mohiam, Imperial Truthsayer — Legendary Creature — Human Advisor {1}{U}{B},
// 1/3 (Secret Lair Drop):
//
//	"{T}: Look at the top card of target player's library. You may put it into
//	 their graveyard.
//	 At the beginning of your end step, you may pay {2}. When you do, exile the
//	 top card of two target players' libraries. For each nonland card exiled
//	 this way, its owner loses life equal to its mana value and it becomes
//	 plotted. (Its owner may cast it as a sorcery on a later turn without
//	 paying its mana cost.)"
//
// The tap ability is Draugr Thought-Thief's look (S82BnsB28DraugrLook) at a
// target player. The end step ability is Anointer of Valor's "may pay, when you
// do": the {2} is paid as it resolves, and only then is the reflexive trigger
// put on the stack (CR 603.12) with its two target players chosen as it goes
// there, each pick a different player. As it resolves the top card of each
// target's library is exiled (the two exiles are independent, so an empty
// library exiles nothing, and a target gone by resolution is skipped, CR 608.2b),
// and each nonland card that reached exile makes its OWNER lose life equal to
// its mana value and become plotted by that owner (CR 702.170a-d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d920b96-eb3e-447a-b6c1-34c3d9cb1403",
		Name:         "Mohiam, Imperial Truthsayer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Look at the top card of target player's library. You may put it into their graveyard.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.S82BnsB28DraugrLook,
		}},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Mohiam, Imperial Truthsayer — you may pay {2}; when you do, exile the top card of two target players' libraries",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Mohiam, Imperial Truthsayer — pay {2}?",
						Then:     mohiamPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// mohiamPaid is the paid branch: the reflexive "when you do". An on-disk key.
var mohiamPaid = effects.PayUnlessBranches("pay-unless/mohiam-imperial-truthsayer-exile",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReflexiveTrigger{
			Label: "Mohiam, Imperial Truthsayer — exile the top card of two target players' libraries",
			Body:  mohiamExileBody,
		}.Apply(ctx)
	}, nil)

// mohiamExileBody exiles the top card of each target player's library. An
// on-disk key.
var mohiamExileBody = game.ReflexiveBody("mohiam-imperial-truthsayer/exile-top", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if err := g.ExileTopKeyedForEffect(t.ID, item.Controller, 1, ctx.ItemTail(mohiamExiled)); err != nil {
			return err
		}
	}
	return nil
}), effects.ConstTargets(func() *game.TargetSpec { return effects.TargetPlayer("two target players").WithCount(2, 2) }))

// mohiamExiled is the exiled card's consequence: each nonland card costs its
// owner life equal to its mana value and becomes plotted. An on-disk key.
var mohiamExiled = effects.CardTail("tail/mohiam-imperial-truthsayer-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.IsLand() {
			continue
		}
		if mv := c.ManaValue(); mv > 0 && c.Owner != uuid.Nil {
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), c.Owner, -mv); err != nil {
				return err
			}
		}
		if err := (effects.PlotExiled{Card: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
