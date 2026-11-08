package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fountainport Bell — Artifact {1} (Bloomburrow, #4467):
//
//	"When this artifact enters, you may search your library for a basic
//	 land card, reveal it, then shuffle and put that card on top.
//	 {1}, Sacrifice this artifact: Draw a card."
//
// The entry trigger is Campus Guide's search: optional, revealed,
// shuffled, then the card goes on top (a declined search does not look
// or shuffle). The draw is Aether Spellbomb's "{1}, Sacrifice this
// artifact: Draw a card", at instant speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ea4c5b8f-a5ff-47dd-8e47-5ae7b21926b4",
		Name:         "Fountainport Bell",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fountainport Bell — search for a basic land card and put it on top", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneLibrary,
					ToTop:     true,
					Limit:     1,
					Optional:  true,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Fountainport Bell — a basic land card",
				}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
