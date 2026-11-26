package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jiang Yanggu (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	+1: Target creature gets +2/+2 until end of turn.
//	−1: If you don't control a creature named Mowu, create Mowu, a legendary 3/3 green Dog creature token.
//	−5: Until end of turn, target creature gains trample and gets +X/+X, where X is the number of lands you control.
//
// {4}{G} Legendary Planeswalker — Yanggu, loyalty 4. The −1 is Venser, Corpse
// Puppet's "if you don't control a creature named" token; the −5 reads X as it
// resolves, as Mirkwood Channeler does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b15c8f43-685d-4901-9cad-47bc6dc6346f",
		Name:            "Jiang Yanggu",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Target creature gets +2/+2 until end of turn.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BoostTargetUntilEOT("Jiang Yanggu — +2/+2", 2, 2),
			},
			{
				Label:   "−1: If you don't control a creature named Mowu, create Mowu, a legendary 3/3 green Dog creature token.",
				Cost:    effects.LoyaltyCost(-1),
				Purpose: game.Purpose{Tokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.YouControlNoCreatureNamed(g, item.Controller, "Mowu") {
						return nil
					}
					return effects.CreateToken{
						Controller: item.Controller,
						Template:   effects.TokenCard("3/3 green Legendary Dog named Mowu"),
						N:          1,
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "−5: Until end of turn, target creature gains trample and gets +X/+X, where X is the number of lands you control.",
				Cost:    effects.LoyaltyCost(-5),
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample")),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					x := len(effects.PermanentsControlledByMatching(g, item.Controller, effects.Land()))
					return effects.GapB125PumpTrample(ctx, id, x, x, "Jiang Yanggu")
				},
			},
		},
	})
}
