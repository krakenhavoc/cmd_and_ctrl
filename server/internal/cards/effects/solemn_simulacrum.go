package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Solemn Simulacrum ("Sad Robot") — 2/2 Artifact Creature — Golem
// with two triggers:
//  1. "When Solemn Simulacrum enters, you may search your library
//     for a basic land card, put that card onto the battlefield
//     tapped, then shuffle."
//  2. "When Solemn Simulacrum dies, you may draw a card."
//
// S14 sandbox simplifications retained:
//   - ETB "may search" is treated as "you do." Empty basics pile
//     no-ops silently (SearchLibrary returns nil on predicate-miss).
//   - The dies-trigger is NOT wired. Triggered-on-leave-battlefield
//     abilities require the listener pipeline that lands in S19.
//     The OnETB hook is the only one available direct-call today.
//
// S17 sub-PR 4: fetched land enters TAPPED via
// SearchLibrary.TappedOnEntry.
func init() {
	Register(Spec{
		OracleID: "00c0543c-2a1f-4425-8283-4062d74a1637",
		Name:     "Solemn Simulacrum",
		OnETB: func(card *game.Card, ctx *Context) error {
			return SearchLibrary{
				Player:        card.Controller,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Reveal:        true,
				Shuffle:       true,
				TappedOnEntry: true,
			}.Apply(ctx)
		},
	})
}
