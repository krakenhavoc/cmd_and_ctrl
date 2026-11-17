package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slimefoot's Survey — Sorcery {4}{G}:
//
//	"Domain — Search your library for up to two land cards that each have
//	 a basic land type, put them onto the battlefield tapped, then
//	 shuffle. Look at the top X cards of your library, where X is the
//	 number of basic land types among lands you control. Put up to one of
//	 them on top of your library and the rest on the bottom of your library
//	 in a random order."
//
// Circuitous Route's search with a "land card that has a basic land type"
// predicate (so a dual land or a Temple Garden qualifies, CR 305.6). X is
// counted only after the lands have entered, so they add to it, and the
// look comes after the shuffle: the tail runs once the search has settled
// (#4253), so the prompt is a restore point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8ace9425-fcdd-45b0-abf0-1a7d601de25e",
		Name:         "Slimefoot's Survey",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:        item.Controller,
				Predicate:     s82DmuB10IsLandWithBasicType,
				Dest:          game.ZoneBattlefield,
				Limit:         2,
				Reveal:        true,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Slimefoot's Survey — up to two land cards that each have a basic land type, onto the battlefield tapped",
				ThenTail:      ctx.Tail(slimefootsSurveyLook),
			}.Apply(ctx)
		},
	})
}

// slimefootsSurveyLook is the second sentence: look at the top X, keep up to
// one on top, the rest under in a random order. An on-disk key.
var slimefootsSurveyLook = CardTail("s82-dmu/slimefoots-survey-look", func(ctx *Context, _ game.TailOutcome) error {
	return s82DmuB10LookKeepOneOnTop(ctx, "Slimefoot's Survey")
})
