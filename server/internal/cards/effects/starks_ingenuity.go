package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "913e6ff0-b522-45a3-9a1e-4a87a515cc1d",
		Name:         "Stark's Ingenuity",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Stark's Ingenuity — you may pay {X}; if you do, draw X cards", starksIngenuityEnters),
			WheneverYouDraw("Stark's Ingenuity — put a +1/+1 counter on enchanted creature", starksIngenuityCounter),
		},
	})
}

func starksIngenuityEnters(g *game.Game, item *game.StackItem) error {
	return MayPayX{
		Cost:  "{X}",
		Label: "Stark's Ingenuity",
		Buys:  "draw X cards",
		Goal:  XAsHighAsYouCan,
		Unit:  game.PayAmountCards,
		Then:  starksIngenuityPaid,
	}.Apply(NewContext(g, item))
}

// starksIngenuityPaid is the paid half, keyed: X rides the prompt.
var starksIngenuityPaid = PayUnlessBranches("pay-unless/starks-ingenuity-draw-x",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		return DrawCards{N: r.Number}.Apply(ctx)
	}, nil)

func starksIngenuityCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := enchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return AddCounter{Target: host, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
