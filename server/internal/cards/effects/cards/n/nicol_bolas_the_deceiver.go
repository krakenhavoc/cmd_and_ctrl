package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nicol Bolas, the Deceiver — {5}{U}{B}{R} Legendary Planeswalker — Bolas (Hour of Devastation), loyalty 5:
//
//	"+3: Each opponent loses 3 life unless that player sacrifices a nonland
//	 permanent of their choice or discards a card.
//	 −3: Destroy target creature. Draw a card.
//	 −11: Nicol Bolas deals 7 damage to each opponent. You draw seven cards."
//
// The +3 is the punisher Perforating Artist uses, asked of each opponent in turn
// order: sacrifice a nonland permanent, discard a card, or lose 3 life.
func init() {
	effects.RegisterPunisher(nicolBolasTheDeceiverPunisher, "nicol-bolas-the-deceiver")
	effects.Register(effects.Spec{
		OracleID:        "524b00ea-0372-40f9-a6a1-76a8497a1ea6",
		Name:            "Nicol Bolas, the Deceiver",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+3: Each opponent loses 3 life unless that player sacrifices a nonland permanent of their choice or discards a card.",
				Cost:  effects.LoyaltyCost(3),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return nicolBolasTheDeceiverPunisher.Start(effects.NewContext(g, item), 1)
				},
			},
			{
				Label:   "−3: Destroy target creature. Draw a card.",
				Purpose: game.Purpose{Draws: 1},
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if err := effects.DestroyFirstLegalCardTarget(g, item); err != nil {
						return err
					}
					return effects.DrawCards{N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "−11: Nicol Bolas deals 7 damage to each opponent. You draw seven cards.",
				Purpose: game.Purpose{Draws: 7},
				Cost:    effects.LoyaltyCost(-11),
				Effect: effects.Do(
					effects.DamageEachOpponent{N: 7},
					effects.DrawCards{N: 7},
				),
			},
		},
	})
}

var nicolBolasTheDeceiverPunisher = effects.PunisherRepeat{Name: "Nicol Bolas, the Deceiver", Life: 3}
