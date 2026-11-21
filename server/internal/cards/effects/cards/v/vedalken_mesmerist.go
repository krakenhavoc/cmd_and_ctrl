package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vedalken Mesmerist — Creature — Vedalken Wizard {1}{U}, 2/1 (Ravnica
// Remastered, #5948):
//
//	"Whenever this creature attacks, target creature an opponent controls gets
//	 -2/-0 until end of turn."
//
// Faerie Duelist's shrink under an attack trigger: the target is chosen as the
// trigger goes on the stack (so with no opposing creature it is removed, CR
// 603.3d) and re-checked as it resolves (CR 608.2b); the -2/-0 ends in the
// cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f507f9ba-a7e9-4e45-86dc-8c60a72a6234",
		Name:         "Vedalken Mesmerist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WheneverThisAttacks("Vedalken Mesmerist — target creature an opponent controls gets -2/-0 until end of turn",
				effects.BoostTargetUntilEOT("Vedalken Mesmerist — -2/-0 until end of turn", -2, 0)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.HarmUntilEOT(0, -2, 0))),
		},
	})
}
