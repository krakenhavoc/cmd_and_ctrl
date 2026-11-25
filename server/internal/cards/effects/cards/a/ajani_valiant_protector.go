package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ajani, Valiant Protector — Legendary Planeswalker — Ajani {4}{G}{W},
// loyalty 4 (Aether Revolt, S82 older sets, #5536):
//
//	"+2: Put two +1/+1 counters on up to one target creature.
//	 +1: Reveal cards from the top of your library until you reveal a creature
//	 card. Put that card into your hand and the rest on the bottom of your
//	 library in a random order.
//	 −11: Put X +1/+1 counters on target creature, where X is your life
//	 total. That creature gains trample until end of turn."
//
// The reveal is one announcement of the whole run; a library with no creature
// card is revealed whole and goes to the bottom in a random order. X is read
// as the ability resolves (CR 608.2h) and the trample is given even when the
// counters are not (a replacement may stop them).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "542423cb-fbfb-4a2f-9f6d-87b273a95e96",
		Name:            "Ajani, Valiant Protector",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Put two +1/+1 counters on up to one target creature.",
				Cost:    effects.LoyaltyCost(2),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect:  effects.TwoPlusOneCountersOnTheTarget,
			},
			{
				Label: "+1: Reveal cards from the top of your library until you reveal a creature card. Put that card into your hand and the rest on the bottom of your library in a random order.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					run, hit := effects.RevealUntil(ctx, player, func(c game.Card) bool { return c.IsCreature() },
						"Ajani, Valiant Protector — revealed until a creature card")
					return effects.TakeFromLibraryToHand{
						Player: player,
						Cards:  run,
						Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
						All:    true,
						Label:  "Ajani, Valiant Protector — put the creature card into your hand",
						Key:    effects.TakeRestBottomRandom,
					}.Apply(ctx)
				},
			},
			{
				Label:   "−11: Put X +1/+1 counters on target creature, where X is your life total. That creature gains trample until end of turn.",
				Cost:    effects.LoyaltyCost(-11),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if p := ctx.Game.PlayerByIDForEffect(ctx.Controller()); p != nil && p.Life > 0 {
						if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: p.Life}).Apply(ctx); err != nil {
							return err
						}
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Ajani, Valiant Protector — trample until end of turn"}.Apply(ctx)
				},
			},
		},
	})
}
