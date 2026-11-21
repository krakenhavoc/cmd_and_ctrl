package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Selhoff Entomber — Creature — Zombie {1}{U}, 1/3 (Innistrad: Crimson Vow):
//
//	"{T}, Discard a creature card: Draw a card."
//
// The discard is a cost (CR 602.2b) named at activation, as Fauna Shaman's
// is; the draw is the whole effect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b97794a-3aaa-45a6-9cad-f1b768615b6d",
		Name:         "Selhoff Entomber",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Discard a creature card: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Discards: 1, Answers: game.AnswerValue},
			Cost: effects.Plus(effects.TapCost(), effects.DiscardCardsMatching(1, "a creature card", func(c game.Card) bool {
				return c.IsCreature()
			})),
			Effect: effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
