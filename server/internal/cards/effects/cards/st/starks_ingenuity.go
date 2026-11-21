package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stark's Ingenuity — Enchantment — Aura {2}{U} (Marvel Super Heroes
// Commander, #4455):
//
//	"Enchant creature you control
//	 When this Aura enters, you may pay {X}. If you do, draw X cards.
//	 Whenever you draw a card, put a +1/+1 counter on enchanted creature."
//
// The enters trigger is MayPayX (X chosen as it resolves, CR 608.2d, #2727)
// with the draw as its keyed paid half. The draw trigger fires once per card
// drawn, including the cards the first ability draws, and puts the counter
// on the creature the Aura is enchanting as it resolves (an Aura that has
// left does nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "913e6ff0-b522-45a3-9a1e-4a87a515cc1d",
		Name:         "Stark's Ingenuity",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Stark's Ingenuity — you may pay {X}; if you do, draw X cards", starksIngenuityEnters),
			effects.WheneverYouDraw("Stark's Ingenuity — put a +1/+1 counter on enchanted creature", effects.PlusOneCounterOnEnchantedCreature),
		},
	})
}

func starksIngenuityEnters(g *game.Game, item *game.StackItem) error {
	return effects.MayPayX{
		Cost:  "{X}",
		Label: "Stark's Ingenuity",
		Buys:  "draw X cards",
		Goal:  effects.XAsHighAsYouCan,
		Unit:  game.PayAmountCards,
		Then:  starksIngenuityPaid,
	}.Apply(effects.NewContext(g, item))
}

// starksIngenuityPaid is the paid half, keyed: X rides the prompt.
var starksIngenuityPaid = effects.PayUnlessBranches("pay-unless/starks-ingenuity-draw-x",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		return effects.DrawCards{N: r.Number}.Apply(ctx)
	}, nil)
