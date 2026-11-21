package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Incisor Glider — Artifact Creature — Phyrexian Construct {1}{W}, 1/3:
//
//	"Flying
//	 Corrupted — Whenever this creature attacks, if an opponent has three or
//	 more poison counters, creatures you control get +1/+1 until end of turn."
//
// An intervening "if" (CR 603.4): checked as the attack is declared and
// again as the trigger resolves. The pump is Overrun's: it affects the
// creatures you control as the ability resolves, not ones that arrive later
// (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e51f89c8-3d7f-4f1a-bc35-67b92db20e10",
		Name:            "Incisor Glider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondCorrupted{}, effects.On(game.EventAttack, effects.AllOf(effects.ThisAttacked, effects.YouAreCorrupted),
				"Corrupted — Incisor Glider — creatures you control get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.Creature(), effects.YouControl()),
						Power:     1,
						Toughness: 1,
						Label:     "Incisor Glider — +1/+1",
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
