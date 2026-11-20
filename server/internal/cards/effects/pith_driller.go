package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pith Driller — Artifact Creature — Phyrexian Horror {4}{B/P}, 2/4 (New Phyrexia, #5575):
//
//	"({B/P} can be paid with either {B} or 2 life.)
//	 When this creature enters, put a -1/-1 counter on target creature."
//
// Dread Tiller's enters trigger: one -1/-1 counter on the target creature
// (CR 122.1). The Phyrexian symbol is paid by the engine at cast (CR
// 107.4f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "62d293a6-4fad-4aaa-af87-62d705e243ca",
		Name:         "Pith Driller",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Pith Driller — put a -1/-1 counter on target creature", cmdr2B25CountersOnTarget(1)),
				Harmful(TargetCreature("target creature")),
			),
		},
	})
}
