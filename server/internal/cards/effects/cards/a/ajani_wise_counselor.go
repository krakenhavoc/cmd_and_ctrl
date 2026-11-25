package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ajaniWiseCounselorMinusNine is the -9: X +1/+1 counters on the target, X the
// controller's life total as it resolves. It sits above init() because the
// read-back lint (TestNoReadBackWithoutAContinuation) reads a file top to
// bottom and this read is not one after the +2's life gain.
func ajaniWiseCounselorMinusNine(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	t, ok := effects.FirstLegalTarget(ctx)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	p := g.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.Life <= 0 {
		return nil
	}
	return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: p.Life}.Apply(ctx)
}

// Ajani, Wise Counselor — Legendary Planeswalker — Ajani {3}{W}{W}, loyalty 5
// (Core Set 2019):
//
//	"+2: You gain 1 life for each creature you control.
//	 −3: Creatures you control get +2/+2 until end of turn.
//	 −9: Put X +1/+1 counters on target creature, where X is your life total."
//
// The +2 counts the creatures its controller controls as it resolves
// (CR 608.2h). The −3 fixes the set of creatures as it resolves (CR 611.2c),
// Black Panther's pump. The −9 reads the controller's life total as it
// resolves and does nothing if the target has gone (CR 608.2b); a life total
// of zero or less puts no counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "341f198b-beb6-4a44-af91-a64a5989c865",
		Name:            "Ajani, Wise Counselor",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: You gain 1 life for each creature you control.",
				Purpose: game.Purpose{LifeGain: 1},
				Cost:    effects.LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.GainLife{Amount: effects.CountYouControl(ctx, effects.Creature())}.Apply(ctx)
				},
			},
			{
				Label: "−3: Creatures you control get +2/+2 until end of turn.",
				Cost:  effects.LoyaltyCost(-3),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Toughness: 2,
						Label: "Ajani, Wise Counselor — creatures you control get +2/+2",
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "−9: Put X +1/+1 counters on target creature, where X is your life total.",
				Cost:    effects.LoyaltyCost(-9),
				Targets: effects.TargetCreature("target creature"),
				Effect:  ajaniWiseCounselorMinusNine,
			},
		},
	})
}
