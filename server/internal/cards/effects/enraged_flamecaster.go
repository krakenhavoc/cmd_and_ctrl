package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enraged Flamecaster — Creature — Elemental Sorcerer {2}{R}, 3/2 (Lorwyn
// Eclipsed, #4458):
//
//	"Reach
//	 Whenever you cast a spell with mana value 4 or greater, this creature
//	 deals 2 damage to each opponent."
//
// Angry Rabble's cast trigger at 2 damage: the spell's mana value is read
// on the stack, so X counts at its announced value. The trigger resolves
// before the spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e35d7886-8c03-471c-b598-118d7eceddb4",
		Name:            "Enraged Flamecaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(ManaValueGE(4), "Enraged Flamecaster — 2 damage to each opponent", Do(DamageEachOpponent{N: 2})),
		},
	})
}
