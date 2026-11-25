package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mu Yanling (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	+2: Target creature can't be blocked this turn.
//	−3: Draw two cards.
//	−10: Tap all creatures your opponents control. You take an extra turn after this one.
//
// {4}{U}{U} Legendary Planeswalker — Yanling, loyalty 5. +2 is Splinter, Radical
// Rat's can't-be-blocked; −10 taps with the Cryptic Command helper, then takes
// the extra turn (Teferi, Master of Time).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "34e1477a-c0fb-4364-bb93-04a85c0c2517",
		Name:            "Mu Yanling",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Target creature can't be blocked this turn.",
				Cost:    effects.LoyaltyCost(2),
				Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.TargetCantBeBlockedThisTurn("Mu Yanling"),
			},
			{
				Label:  "−3: Draw two cards.",
				Cost:   effects.LoyaltyCost(-3),
				Effect: effects.Do(effects.DrawCards{N: 2}),
			},
			{
				Label:   "−10: Tap all creatures your opponents control. You take an extra turn after this one.",
				Cost:    effects.LoyaltyCost(-10),
				Purpose: game.Purpose{ExtraTurns: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					if err := effects.TapAllCreaturesYourOpponentsControl(g, item); err != nil {
						return err
					}
					return effects.TakeExtraTurn{}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
