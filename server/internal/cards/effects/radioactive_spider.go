package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Radioactive Spider — Creature — Spider {G}, 1/1 (Marvel's Spider-Man, #4460):
//
//	"Reach, deathtouch
//	 Fateful Bite — {2}, Sacrifice this creature: Search your library for
//	 a Spider Hero card, reveal it, put it into your hand, then shuffle.
//	 Activate only as a sorcery."
//
// The sacrifice is a cost, so the search happens even though the Spider
// is gone by resolution. "A Spider Hero card" is a stated quality: the
// search may find nothing (CR 701.23b), and the pick must carry both
// subtypes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f15bd5c0-99ff-42c6-8eed-48de6df81b61",
		Name:            "Radioactive Spider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "deathtouch"},
		Activated: []ActivatedAbility{{
			Label:        "Fateful Bite — {2}, Sacrifice this creature: Search your library for a Spider Hero card, reveal it, put it into your hand, then shuffle. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}"), SacrificeThis()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return c.HasSubtypeWord("Spider") && c.HasSubtypeWord("Hero")
					},
					Dest:    game.ZoneHand,
					Limit:   1,
					Reveal:  true,
					Shuffle: true,
					Reason:  "Radioactive Spider — search for a Spider Hero card",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
