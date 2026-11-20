package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grim Physician — Creature — Zombie {B}, 1/1 (Theros Beyond Death, #5514):
//
//	"When this creature dies, target creature an opponent controls
//	 gets -1/-1 until end of turn."
//
// Front Porch Sentries' dies trigger: the -1/-1 lands on a creature an
// opponent controls, until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "381539ce-fc05-4855-80da-b0807a7c07c2",
		Name:         "Grim Physician",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Grim Physician — target creature an opponent controls gets -1/-1 until end of turn",
				boostTargetUntilEOT("Grim Physician — -1/-1", -1, -1)),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
