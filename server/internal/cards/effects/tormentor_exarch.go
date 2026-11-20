package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tormentor Exarch — Creature — Phyrexian Cleric {3}{R}, 2/2 (New Phyrexia, #5575):
//
//	"When this creature enters, choose one —
//	 • Target creature gets +2/+0 until end of turn.
//	 • Target creature gets -0/-2 until end of turn."
//
// A modal enters trigger (Adaptive Sporesinger's shape): the bullet is
// chosen as the trigger goes on the stack (CR 603.3c), and each bullet
// targets a creature (CR 603.3d).
//
// No simplification.
func init() {
	trigger := WhenThisEnters("Tormentor Exarch — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeDoing("Target creature gets +2/+0 until end of turn.", TargetCreature("target creature"),
			BoostTheModesTarget(2, 0, "Tormentor Exarch — +2/+0")),
		ModeDoing("Target creature gets -0/-2 until end of turn.", TargetCreature("target creature"),
			BoostTheModesTarget(0, -2, "Tormentor Exarch — -0/-2")),
	)
	Register(Spec{
		OracleID:     "ebc52f9a-ba0c-48ed-bb04-ed086efeb060",
		Name:         "Tormentor Exarch",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
