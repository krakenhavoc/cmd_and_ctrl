package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "495054c8-d21d-4b89-9853-bb299d30ef63",
		Name:            "Hulk, Always Angry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{effects.AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisEnters("Hulk — destroy all artifacts", func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyAllMatching{Match: effects.Artifact()}.Apply(effects.NewContext(g, item))
			}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepDestroy}}),
		},
	})
}
