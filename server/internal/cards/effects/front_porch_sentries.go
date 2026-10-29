package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Front Porch Sentries — Creature — Goblin Soldier {1}{B}, 2/2 (The Hobbit):
//
//	"When this creature dies, target creature an opponent controls gets
//	 -1/-1 until end of turn."
//
// A dies trigger with a target chosen as it goes on the stack (CR 603.3d),
// limited to creatures an opponent controls. The shrink resolves on
// whatever it is by then; a target that has left or become illegal makes
// the whole ability fizzle (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b39f03cf-3d1a-42ad-903f-e3b3542878b0",
		Name:         "Front Porch Sentries",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Front Porch Sentries — target creature an opponent controls gets -1/-1 until end of turn",
				boostTargetUntilEOT("Front Porch Sentries — -1/-1", -1, -1)),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
