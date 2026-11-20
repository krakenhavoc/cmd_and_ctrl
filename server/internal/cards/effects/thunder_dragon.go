package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thunder Dragon — Creature — Dragon {5}{R}{R}, 5/5 (Starter 1999,
// #5620):
//
//	"Flying
//	 When this creature enters, it deals 3 damage to each creature
//	 without flying."
//
// Crater Hellion's enters sweep, narrowed to creatures without flying as
// the trigger resolves (Harbinger of the Hunt's filter). The Dragon is
// the source; it has flying, so it spares itself, and one that has lost
// flying by then is dealt the damage too, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "70d9579f-c8a6-4406-afe1-9291fe5ee736",
		Name:            "Thunder Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		// ADR 0126 §6: only creatures without flying.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3, Partial: true}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Thunder Dragon — 3 damage to each creature without flying", func(g *game.Game, item *game.StackItem) error {
				return damageEachMatching(NewContext(g, item), And(Creature(), WithoutKeyword("flying")), 3)
			}),
		},
	})
}
