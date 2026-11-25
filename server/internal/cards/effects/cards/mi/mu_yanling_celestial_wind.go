package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mu Yanling, Celestial Wind — Legendary Planeswalker — Yanling {4}{U}{U},
// loyalty 5 (Core Set 2020, #5518):
//
//	"+1: Until your next turn, up to one target creature gets -5/-0.
//	 −3: Return up to two target creatures to their owners' hands.
//	 −7: Creatures you control with flying get +5/+5 until end of turn."
//
// The +1 is a scoped effect lasting until the activator's next turn begins
// (CR 611.2). The −3 is Waterwhirl's simultaneous bounce. The −7 locks the set
// of creatures you control that have flying as it resolves (CR 611.2c), read
// through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5f39accf-62fd-407a-a147-781210569d33",
		Name:            "Mu Yanling, Celestial Wind",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Until your next turn, up to one target creature gets -5/-0.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect:  effects.OldB155MuYanlingPlusOne,
			},
			{
				Label:   "−3: Return up to two target creatures to their owners' hands.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("up to two target creatures").WithCount(0, 2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					g.BounceCardsToHandForEffect(effects.LegalTargetCardIDs(effects.NewContext(g, item)))
					return nil
				},
			},
			{
				Label: "−7: Creatures you control with flying get +5/+5 until end of turn.",
				Cost:  effects.LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.Creature(), effects.YouControl(), effects.HasKeyword("flying")),
						Power:     5,
						Toughness: 5,
						Label:     "Mu Yanling, Celestial Wind — +5/+5",
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
