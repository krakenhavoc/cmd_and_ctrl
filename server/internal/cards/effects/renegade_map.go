package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f8506acb-3fc5-4fdf-8304-586843db6cfd",
		Name:         "Renegade Map",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this artifact: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Renegade Map — a basic land card",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
