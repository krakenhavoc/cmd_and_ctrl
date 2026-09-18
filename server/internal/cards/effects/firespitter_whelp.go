package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Firespitter Whelp — 2/2 Creature — Dragon for {2}{R} (EDHREC rank
// 4031):
//
//	"Flying
//	 Whenever you cast a noncreature or Dragon spell, this creature
//	 deals 1 damage to each opponent."
//
// A three-mana flier that turns a spellslinger or Dragon deck into a
// clock on the whole table. It is in the batch because the trigger
// condition is a genuine OR of two different tests — a card type and
// a creature type — rather than the single test every other
// "whenever you cast" card in the catalog makes.
//
// The two halves overlap in exactly the place you would expect: a
// Dragon creature spell is a creature spell, so the noncreature test
// fails and the Dragon test carries it. A Dragon ARTIFACT or a Dragon
// enchantment (a Dragon-typed Saga, an Ur-Dragon Aura) satisfies both
// and still triggers once — this is one ability with a disjunctive
// condition, not two abilities.
//
// "You cast" is the Whelp's controller only; an opponent's Dragons do
// nothing. The damage is the Whelp's, so it is noncombat damage from a
// red creature source (Torbran sees it) and every opponent takes it
// individually — a player who has left the game takes none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1cad92ce-55c8-4b78-8d8c-56645ef8e6ee",
		Name:            "Firespitter Whelp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Noncreature(), HasSubtype("Dragon")),
				"Firespitter Whelp — 1 damage to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					return damageToEachOpponent(g, item, 1)
				}),
		},
	})
}
