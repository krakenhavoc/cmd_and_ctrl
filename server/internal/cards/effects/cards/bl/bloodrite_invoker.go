package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodrite Invoker — Creature — Vampire Shaman {2}{B}, 3/1 (Duel Decks:
// Zendikar vs. Eldrazi, #5548):
//
//	"{8}: Target player loses 3 life and you gain 3 life."
//
// An illegal target does nothing, and the life is not gained (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2dd3a6f7-65a9-4157-9ad9-13a885ebe927",
		Name:         "Bloodrite Invoker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{8}: Target player loses 3 life and you gain 3 life.",
			Cost:    effects.ManaCost("{8}"),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.TargetPlayerLosesLifeYouGain(3)(item, effects.NewContext(g, item))
			},
		}},
	})
}
