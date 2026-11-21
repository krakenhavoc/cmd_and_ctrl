package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Renegade Map — Artifact {1} (Aether Revolt,
// #5536):
//
//	"This artifact enters tapped.
//	 {T}, Sacrifice this artifact: Search your library for a basic land
//	 card, reveal it, put it into your hand, then shuffle."
//
// Worn Powerstone's enters-tapped replacement and Horizon Seeker's
// basic land search to hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8506acb-3fc5-4fdf-8304-586843db6cfd",
		Name:         "Renegade Map",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this artifact: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Renegade Map — a basic land card",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
