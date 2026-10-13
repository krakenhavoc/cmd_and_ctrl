package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Delivery Moogle — Creature — Moogle {3}{W}, 3/2:
//
//	"Flying
//	 When this creature enters, search your library and/or graveyard for
//	 an artifact card with mana value 2 or less, reveal it, and put it
//	 into your hand. If you search your library this way, shuffle."
//
// Not a "may": the search is compulsory, and an artifact card with mana
// value 2 or less in the graveyard is a public match, so the engine
// makes the searcher take one while any is there (CR 701.23b is a
// hidden-zone rule, #3217). Taking a graveyard card skips the shuffle;
// taking a library artifact, or none, shuffles.
func init() {
	Register(Spec{
		OracleID:        "60043b87-6008-4cfa-9d8f-6ee6e3f70766",
		Name:            "Delivery Moogle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, b06SelfETB, "Delivery Moogle — search for an artifact card with mana value 2 or less",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     func(c game.Card) bool { return c.IsArtifact() && c.ManaValue() <= 2 },
						Dest:          game.ZoneHand,
						Limit:         1,
						Reveal:        true,
						Shuffle:       true,
						AlsoGraveyard: true,
						Reason:        "Delivery Moogle — an artifact card with mana value 2 or less",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
