package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rhystic Tutor — Sorcery {2}{B}:
//
//	"Unless any player pays {2}, search your library for a card, put
//	 that card into your hand, then shuffle."
//
// ADR 0150 (#1903): the {2} is offered to every player in APNAP order
// from the active player, the caster included, until one pays
// (AnyPlayerPays). The search is the rest of the spell, so the offer
// holds the step until it is answered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c6da9a9-6395-460b-8150-ea46d7299dc2",
		Name:         "Rhystic Tutor",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.AnyPlayerPays{
				Cost:     "{2}",
				Question: "Rhystic Tutor — pay {2} to stop the search?",
				Then:     rhysticTutorSearches,
			}.Apply(ctx)
		},
	})
}

// rhysticTutorSearches is "if no one does", keyed (#4264).
var rhysticTutorSearches = effects.AnyPlayerConsequences("any-player/rhystic-tutor-search",
	nil,
	func(ctx *effects.Context, _ game.AnyPlayerAnswer) error {
		return effects.SearchLibrary{
			Player:  ctx.Controller(),
			Dest:    game.ZoneHand,
			Limit:   1,
			Shuffle: true,
			Reason:  "Rhystic Tutor — search your library for a card",
		}.Apply(ctx)
	},
	nil)
