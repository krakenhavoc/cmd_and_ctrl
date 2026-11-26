package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stampeding Elk Herd — Creature — Elk {3}{G}{G}, 5/5 (Dragons of Tarkir, S82
// older sets):
//
//	"Formidable — Whenever this creature attacks, if creatures you control have
//	 total power 8 or greater, creatures you control gain trample until end of
//	 turn."
//
// Sabertooth Outrider's formidable attack trigger (an intervening "if", CR 603.4)
// with a group grant: the creatures you control as it resolves gain trample
// until end of turn, the set fixed then (CR 611.2c), so a creature that arrives
// later does not have it.
//
// No simplification.
func init() {
	const label = "Stampeding Elk Herd — formidable: creatures you control gain trample until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "fc1020b7-086e-41d6-a9ae-1c6c72cf820c",
		Name:         "Stampeding Elk Herd",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB324FormidableAttack(label, func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"trample"},
					Label:    label,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
