package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Huatli, Warrior Poet — Legendary Planeswalker — Huatli {3}{R}{W},
// loyalty 3 (Ixalan, #5532):
//
//	"+2: You gain life equal to the greatest power among creatures you
//	     control.
//	 0: Create a 3/3 green Dinosaur creature token with trample.
//	 −X: Huatli deals X damage divided as you choose among any number of
//	     target creatures. Creatures dealt damage this way can't block this
//	     turn."
//
// The +2 reads the greatest power on the layered board as it resolves, and
// gains nothing at 0 or less. The −X is a loyalty cost of X (CR 606.6) that
// is also the amount divided among the chosen creatures (CR 601.2d, Rolling
// Thunder's shape): each takes its announced share as one damage instance,
// a target that stopped being legal takes nothing, and only a creature that
// was really dealt damage (not one whose damage was prevented) can't block
// this turn (Ballista Wielder's rider).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3ec889bc-f4d7-492a-91c3-d3691eb90eac",
		Name:            "Huatli, Warrior Poet",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		XMatters:        true,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: You gain life equal to the greatest power among creatures you control.",
				Cost:    effects.LoyaltyCost(2),
				Purpose: game.Purpose{},
				Effect: func(g *game.Game, item *game.StackItem) error {
					n := effects.B42GreatestPowerControlledBy(g, item.Controller)
					if n <= 0 {
						return nil
					}
					return effects.GainLife{Player: item.Controller, Amount: n}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "0: Create a 3/3 green Dinosaur creature token with trample.",
				Cost:    effects.LoyaltyCost(0),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Dinosaur with trample"), N: 1}),
			},
			{
				Label:   "−X: Huatli deals X damage divided as you choose among any number of target creatures. Creatures dealt damage this way can't block this turn.",
				Cost:    effects.LoyaltyMinusX(),
				Targets: effects.TargetCreature("any number of target creatures").WithCount(0, 0).Dividing(effects.DivideX()),
				Purpose: game.Purpose{},
				Effect:  effects.OldB228HuatliDividedDamage,
			},
		},
	})
}
