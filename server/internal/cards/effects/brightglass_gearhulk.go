package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brightglass Gearhulk — Artifact Creature — Construct {G}{G}{W}{W}, 4/4
// (Aetherdrift, #4464):
//
//	"First strike, trample
//	 When this creature enters, you may search your library for up to two
//	 artifact, creature, and/or enchantment cards with mana value 1 or
//	 less, reveal them, put them into your hand, then shuffle."
//
// Trinket Mage's tutor with a wider type list and a count of two. "You
// may search" is Optional: declining skips the search and the shuffle.
// Mana value is the card's in the library, so a card with {X} in its cost
// counts X as 0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0c7019a7-3cca-481a-9af0-abedb89c15ad",
		Name:            "Brightglass Gearhulk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Brightglass Gearhulk — search for up to two artifact, creature, and/or enchantment cards with mana value 1 or less",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player: item.Controller,
						Predicate: func(c game.Card) bool {
							return (c.IsArtifact() || c.IsCreature() || c.IsEnchantment()) && c.ManaValue() <= 1
						},
						Dest:     game.ZoneHand,
						Limit:    2,
						Reveal:   true,
						Shuffle:  true,
						Optional: true,
						Reason:   "Brightglass Gearhulk — up to two artifact, creature, and/or enchantment cards with mana value 1 or less",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
