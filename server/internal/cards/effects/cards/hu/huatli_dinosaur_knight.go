package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Huatli, Dinosaur Knight — Legendary Planeswalker — Huatli {4}{R}{W},
// loyalty 4 (Ixalan, #5532):
//
//	"+2: Put two +1/+1 counters on up to one target Dinosaur you control.
//	 −3: Target Dinosaur you control deals damage equal to its power to
//	     target creature you don't control.
//	 −7: Dinosaurs you control get +4/+4 until end of turn."
//
// The +2 takes an "up to one" target, so none is fine. The −3 is Bite
// Down's one-sided bite over two clauses chosen as the ability goes on the
// stack: both must still be legal as it resolves, or nothing happens
// (CR 608.2b), and the damage is the Dinosaur's power as it is then. The −7
// pumps the Dinosaurs you control as it resolves (CR 611.2c fixes the set).
//
// No simplification.
func init() {
	dinosaurYouControl := func(label string) *game.TargetSpec {
		return effects.TargetCreature(label, effects.Subtype("Dinosaur"), effects.YouControl())
	}
	effects.Register(effects.Spec{
		OracleID:        "04aa210a-235f-4e07-87d1-0d28cdf6888b",
		Name:            "Huatli, Dinosaur Knight",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Put two +1/+1 counters on up to one target Dinosaur you control.",
				Cost:    effects.LoyaltyCost(2),
				Targets: dinosaurYouControl("up to one target Dinosaur you control").WithCount(0, 1),
				Purpose: game.Purpose{},
				Effect:  effects.TwoPlusOneCountersOnFirstLegalTarget,
			},
			{
				Label: "−3: Target Dinosaur you control deals damage equal to its power to target creature you don't control.",
				Cost:  effects.LoyaltyCost(-3),
				Targets: effects.Clauses(
					dinosaurYouControl("target Dinosaur you control"),
					effects.TargetCreature("target creature you don't control", effects.OpponentControls())),
				Purpose: game.Purpose{},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.OneSidedBite(item, effects.NewContext(g, item))
				},
			},
			{
				Label:   "−7: Dinosaurs you control get +4/+4 until end of turn.",
				Cost:    effects.LoyaltyCost(-7),
				Purpose: game.Purpose{},
				Effect: effects.Do(effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.Subtype("Dinosaur"), effects.YouControl()), Power: 4, Toughness: 4,
					Label: "Huatli, Dinosaur Knight — +4/+4",
				}),
			},
		},
	})
}
