package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shrine Steward — Artifact Creature — Construct {5}, 3/2 (Kamigawa: Neon
// Dynasty, #5493):
//
//	"When this creature enters, you may search your library for an Aura or
//	 Shrine card, reveal it, put it into your hand, then shuffle."
//
// Stoneforge Mystic's enters trigger over "an Aura or Shrine card": the
// "you may" is the search prompt's decline, which skips the shuffle too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "993a91fa-84ed-4688-9884-e03cbad6c791",
		Name:         "Shrine Steward",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Shrine Steward — search for an Aura or Shrine card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.HasSubtype("Aura") || c.HasSubtype("Shrine") },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Shrine Steward — an Aura or Shrine card, revealed, to hand",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
