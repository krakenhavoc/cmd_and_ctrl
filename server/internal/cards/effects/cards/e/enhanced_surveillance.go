package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enhanced Surveillance — Enchantment {1}{U}:
//
//	"You may look at an additional two cards each time you surveil.
//	 Exile this enchantment: Shuffle your graveyard into your library."
//
// The first ability is SurveilPlusOne's replacement of the surveil keyword
// action, with +2 and a "may" asked each time you would surveil
// (cmdr2B27SurveilPlusTwoMay). The second is Feldon's Cane's effect with the
// exile as its only cost, paid from the battlefield so the enchantment is not
// shuffled in with the graveyard it resets.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfb5104b-01e3-4d37-ac93-1dc342121531",
		Name:         "Enhanced Surveillance",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.Cmdr2B27SurveilPlusTwoMay("Enhanced Surveillance — you may look at an additional two cards each time you surveil"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Exile this enchantment: Shuffle your graveyard into your library.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ExileThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.FinaleShuffleGraveyardIntoLibrary(effects.NewContext(g, item), item.Controller)
			},
		}},
	})
}
