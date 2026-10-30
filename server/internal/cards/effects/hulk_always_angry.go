package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hulk, Always Angry — Legendary Creature — Gamma Berserker Hero {5}{R}{R},
// 9/9 (Marvel Super Heroes Commander):
//
//	"Trample
//	 When Hulk enters, destroy all artifacts.
//	 Hulk attacks each combat if able."
//
// The enter trigger is Shatterstorm's sweep (every artifact on the table, the
// controller's included), declared as a sweep so the board-wipe audit sees
// it. The attack requirement is a static (CR 508.1d), refused at declaration.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "495054c8-d21d-4b89-9853-bb299d30ef63",
		Name:            "Hulk, Always Angry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(WhenThisEnters("Hulk — destroy all artifacts", func(g *game.Game, item *game.StackItem) error {
				return DestroyAllMatching{Match: Artifact()}.Apply(NewContext(g, item))
			}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepDestroy}}),
		},
	})
}
