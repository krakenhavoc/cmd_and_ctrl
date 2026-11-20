package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Novablast Wurm — Creature — Wurm {3}{G}{G}{W}{W}, 7/7 (Worldwake, #5582):
//
//	"Whenever this creature attacks, destroy all other creatures."
//
// An attack trigger with Deathbringer Regent's sweep: every creature
// but the Wurm is destroyed as it resolves, attackers and the
// controller's own included. The sweep spares the Wurm, so it is
// declared partial.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c3aa2554-1cea-45e7-8fc7-1c09bd82fadb",
		Name:         "Novablast Wurm",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Novablast Wurm — destroy all other creatures", func(g *game.Game, item *game.StackItem) error {
				return DestroyAllMatching{Match: And(Creature(), OtherThan(item.SourceCardID))}.Apply(NewContext(g, item))
			}),
		},
	})
}
