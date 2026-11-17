package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disciples of Gix — Creature — Phyrexian Human {4}{B}{B}, 4/4 (The
// Brothers' War, #5482):
//
//	"When this creature enters, search your library for up to three
//	 artifact cards, put them into your graveyard, then shuffle."
//
// Buried Alive's search with the quality changed: a stated quality, so the
// searcher may find fewer (CR 701.23b), and the cards go to the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb7e451b-a9d2-4f50-9651-b5df34e34d99",
		Name:         "Disciples of Gix",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{SelfMillTutor: 3},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Disciples of Gix — search for up to three artifact cards and put them into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsArtifact() },
						Dest:      game.ZoneGraveyard,
						Limit:     3,
						Shuffle:   true,
						Reason:    "Disciples of Gix — put up to three artifact cards into your graveyard",
					}.Apply(ctx)
				}),
		},
	})
}
