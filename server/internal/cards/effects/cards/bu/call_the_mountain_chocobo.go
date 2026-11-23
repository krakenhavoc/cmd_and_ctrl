package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Call the Mountain Chocobo — Sorcery {3}{R} (Final Fantasy, #4462):
//
//	"Search your library for a Mountain card, reveal it, put it into your
//	 hand, then shuffle. Create a 2/2 green Bird creature token with
//	 "Whenever a land you control enters, this token gets +1/+0 until end
//	 of turn."
//	 Flashback {5}{R} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Oliphaunt's mountaincycling search as a spell: "a Mountain card" is
// any land with the Mountain type, not only a basic one. The searcher
// may find nothing (CR 701.23b), and the library is shuffled either
// way. The Bird is the set's
// Chocobo Bird (chocoboBirdToken), made after the search is answered
// (CR 608.2c). Flashback is the shared constructor (CR 702.34a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "a0dd3eb4-f6f1-4bfb-935c-032dd5af3811",
		Name:             "Call the Mountain Chocobo",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Tutors: 1, Tokens: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{5}{R}")},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(callTheMountainChocoboSearch),
			effects.CreateToken{Template: effects.ChocoboBirdToken(), N: 1},
		),
	})
}

// callTheMountainChocoboSearch is "search your library for a Mountain
// card, reveal it, put it into your hand, then shuffle".
func callTheMountainChocoboSearch(ctx *effects.Context) error {
	return effects.SearchLibrary{
		Player:    ctx.Controller(),
		Predicate: effects.IsLandWithSubtype("Mountain"),
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Call the Mountain Chocobo — a Mountain card, to your hand",
	}.Apply(ctx)
}
