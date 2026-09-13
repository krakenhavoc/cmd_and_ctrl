package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rune-Scarred Demon — Creature — Demon {5}{B}{B}, 6/6 (EDHREC rank
// 1296):
//
//	"Flying
//	 When this creature enters, search your library for a card, put it
//	 into your hand, then shuffle."
//
// Demonic Tutor on a 6/6 flier. The ETB is a real trigger with a
// response window; the search is unfiltered and unrevealed, through
// the S22 chooser, exactly as the Tutor's is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "14aefe99-fb60-4f1c-a71f-7ffbe94c8b13",
		Name:            "Rune-Scarred Demon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Rune-Scarred Demon — search your library for a card",
					func(g *game.Game, item *game.StackItem) error {
						return SearchLibrary{
							Player:  item.Controller,
							Dest:    game.ZoneHand,
							Limit:   1,
							Shuffle: true,
							Reason:  "Rune-Scarred Demon — search your library for a card",
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
