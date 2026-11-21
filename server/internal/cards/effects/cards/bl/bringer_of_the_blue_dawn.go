package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bringer of the Blue Dawn — Creature — Bringer {7}{U}{U}, 5/5:
//
//	"You may pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost.
//	 Trample
//	 At the beginning of your upkeep, you may draw two cards."
//
// The alternative cost is bringerAlternativeCost (CR 118.9). The upkeep
// draw is a "you may" asked as the trigger RESOLVES (MayChoice, CR
// 608.2d), not when it goes on the stack, so an opponent's answer to the
// trigger cannot be played around by having already said yes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "7d4676fa-4bfe-43b4-9241-c2a33214852a",
		Name:             "Bringer of the Blue Dawn",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{effects.BringerAlternativeCost()},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Bringer of the Blue Dawn — you may draw two cards", func(g *game.Game, item *game.StackItem) error {
				return effects.MayChoice{
					Question: "Bringer of the Blue Dawn — draw two cards?",
					Draws:    2,
					Then:     effects.MayDrawCards,
					Word:     "2",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
