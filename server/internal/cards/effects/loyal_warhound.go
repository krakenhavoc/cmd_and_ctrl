package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loyal Warhound — "Vigilance. When this creature enters, if an
// opponent controls more lands than you, search your library for a
// basic Plains card, put it onto the battlefield tapped, then
// shuffle."
//
// The intervening-if (CR 603.4) is checked twice in paper: once
// when the trigger would go on the stack and again on resolution.
// AppliesTo covers the first check; the Effect re-checks so a land
// drop in response correctly fizzles the fetch.
//
// Reviewed against the oracle text for #1306: the search is not a
// "may", so it is not Optional; "a basic Plains card" admits a
// Snow-Covered Plains (IsBasicLandWithSubtype reads the Basic
// supertype) and not a nonbasic Plains such as Hallowed Fountain; the
// Plains enters tapped; the library is shuffled. Lands are counted
// among the permanents on the battlefield, so a phased-out land
// (held off the battlefield slice, ADR 0084) is not counted.
func init() {
	Register(Spec{
		OracleID:        "cc6a83c7-e645-4a53-9550-be79b42cd851",
		Name:            "Loyal Warhound",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID != source.InstanceID {
					return false
				}
				return anOpponentHasMoreLands(g, source.Controller)
			}, "Loyal Warhound — search for a basic Plains", func(g *game.Game, item *game.StackItem) error {
				if !anOpponentHasMoreLands(g, item.Controller) {
					return nil
				}
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     IsBasicLandWithSubtype("plains"),
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Loyal Warhound — a basic Plains",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
