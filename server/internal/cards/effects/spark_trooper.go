package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spark Trooper — Creature — Elemental Soldier {1}{R}{R}{W}, 6/1 (Gatecrash, #5565):
//
//	"Trample, lifelink, haste
//	 At the beginning of the end step, sacrifice this creature."
//
// Blistering Firecat's end-step sacrifice: it triggers at the beginning
// of every end step, not only its controller's. The keywords are the
// engine's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3c82389c-7442-4599-89a0-fb2482e727a3",
		Name:            "Spark Trooper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "lifelink", "haste"},
		Triggered: []game.TriggeredAbility{
			AtEachStep(game.StepEnd, "Spark Trooper — sacrifice it",
				func(g *game.Game, item *game.StackItem) error {
					return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
	})
}
