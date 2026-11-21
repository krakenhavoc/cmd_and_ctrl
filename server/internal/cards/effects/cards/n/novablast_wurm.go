package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c3aa2554-1cea-45e7-8fc7-1c09bd82fadb",
		Name:         "Novablast Wurm",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Novablast Wurm — destroy all other creatures", func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.OtherThan(item.SourceCardID))}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
