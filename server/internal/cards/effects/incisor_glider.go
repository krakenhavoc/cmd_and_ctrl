package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "e51f89c8-3d7f-4f1a-bc35-67b92db20e10",
		Name:            "Incisor Glider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondCorrupted{}, On(game.EventAttack, AllOf(ThisAttacked, YouAreCorrupted),
				"Corrupted — Incisor Glider — creatures you control get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match:     And(Creature(), YouControl()),
						Power:     1,
						Toughness: 1,
						Label:     "Incisor Glider — +1/+1",
					}.Apply(NewContext(g, item))
				})),
		},
	})
}
