package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "381539ce-fc05-4855-80da-b0807a7c07c2",
		Name:         "Grim Physician",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Grim Physician — target creature an opponent controls gets -1/-1 until end of turn",
				effects.BoostTargetUntilEOT("Grim Physician — -1/-1", -1, -1)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
