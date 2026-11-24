package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ajani, Mentor of Heroes — Legendary Planeswalker — Ajani {3}{G}{W},
// loyalty 4:
//
//	"+1: Distribute three +1/+1 counters among one, two, or three target
//	 creatures you control.
//	 +1: Look at the top four cards of your library. You may reveal an
//	 Aura, creature, or planeswalker card from among them and put it into
//	 your hand. Put the rest on the bottom of your library in any order.
//	 −8: You gain 100 life."
//
// The first +1 divides as it is put on the stack (CR 601.2d, CR 115.6):
// each of the one to three targets gets at least one counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a612bd38-069f-4bdc-9aa6-c172435e3841",
		Name:            "Ajani, Mentor of Heroes",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Distribute three +1/+1 counters among one, two, or three target creatures you control.",
				Cost:  effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("one, two, or three target creatures you control", effects.YouControl()).
					WithCount(1, 3).Dividing(effects.Divide(3)),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error { return effects.PutDividedCounters(ctx, game.CounterPlusOne) })),
			},
			{
				Label: "+1: Look at the top four cards of your library. You may reveal an Aura, creature, or planeswalker card from among them and put it into your hand. Put the rest on the bottom of your library in any order.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 4),
						Match: effects.Or(effects.Creature(), effects.Planeswalker(),
							func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsAura() }),
						Max:      1,
						Optional: true,
						Label:    "Ajani, Mentor of Heroes — you may reveal an Aura, creature, or planeswalker card from among them and put it into your hand",
						Key:      effects.TakeRevealRestBottomAny,
					}.Apply(ctx)
				},
			},
			{
				Label:  "−8: You gain 100 life.",
				Cost:   effects.LoyaltyCost(-8),
				Effect: effects.Do(effects.GainLife{Amount: 100}),
			},
		},
	})
}
