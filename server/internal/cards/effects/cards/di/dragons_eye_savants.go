package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dragon's Eye Savants — Khans of Tarkir (S82 OLD b345, #5554):
//
//	Creature — Human Wizard {1}{U}, 0/6: Morph—Reveal a blue card in your hand. When
//	this creature is turned face up, look at target opponent's hand.

func init() {
	effects.Register(effects.Spec{
		OracleID:         "ca588f4e-4d2f-4aae-84de-be9862e13f1c",
		Name:             "Dragon's Eye Savants",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.MorphReveal("U")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisIsTurnedFaceUp("Dragon's Eye Savants — look at target opponent's hand", func(g *game.Game, item *game.StackItem) error {
					for _, t := range effects.NewContext(g, item).LegalTargets() {
						if t.Kind == game.TargetPlayer {
							g.LookAtHandForEffect(item.Controller, t.ID)
							break
						}
					}
					return nil
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
