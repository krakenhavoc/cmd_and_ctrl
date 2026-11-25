package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vivien of the Arkbow — Legendary Planeswalker — Vivien {4}{G}{G}, loyalty 5
// (Core Set 2019):
//
//	"+2: Put two +1/+1 counters on up to one target creature.
//	 −3: Target creature you control deals damage equal to its power to target
//	 creature you don't control.
//	 −9: Creatures you control get +4/+4 and gain trample until end of turn."
//
// The +2 is Basri Ket's up-to-one target. The −3 is Vivien, Arkbow Ranger's
// one-sided bite over two clauses (damage only when both targets are still
// legal, CR 608.2b). The −9 fixes the set of creatures as it resolves
// (CR 611.2c) for both the pump and the keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "34fe53c2-8909-4786-b5c0-331de997ff81",
		Name:            "Vivien of the Arkbow",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Put two +1/+1 counters on up to one target creature.",
				Cost:    effects.LoyaltyCost(2),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
				},
			},
			{
				Label: "−3: Target creature you control deals damage equal to its power to target creature you don't control.",
				Cost:  effects.LoyaltyCost(-3),
				Targets: effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
				),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.OneSidedBite(item, effects.NewContext(g, item))
				},
			},
			{
				Label: "−9: Creatures you control get +4/+4 and gain trample until end of turn.",
				Cost:  effects.LoyaltyCost(-9),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					match := effects.And(effects.Creature(), effects.YouControl())
					if err := (effects.BoostUntilEOT{Match: match, Power: 4, Toughness: 4,
						Label: "Vivien of the Arkbow — creatures you control get +4/+4"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: match, Keywords: []string{"trample"},
						Label: "Vivien of the Arkbow — creatures you control gain trample"}.Apply(ctx)
				},
			},
		},
	})
}
