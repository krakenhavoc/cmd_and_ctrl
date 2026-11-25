package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whisper Squad — Creature — Human Soldier {B}, 1/1 (Ikoria: Lair of Behemoths, #5512):
//
//	"{1}{B}: Search your library for a card named Whisper Squad, put it onto
//	 the battlefield tapped, then shuffle."
//
// Naming a quality the engine can't verify, the search may find nothing
// (CR 701.23b) and still shuffles; the card enters tapped. Nothing is
// revealed: the card does not say to.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6b27729-fea5-46c9-8624-07ba291791df",
		Name:         "Whisper Squad",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: Search your library for a card named Whisper Squad, put it onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{1}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:        item.Controller,
					Predicate:     func(c game.Card) bool { return game.CardNameMatches(c, "Whisper Squad") },
					Dest:          game.ZoneBattlefield,
					TappedOnEntry: true,
					Limit:         1,
					Shuffle:       true,
					Reason:        "Whisper Squad — a card named Whisper Squad",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
