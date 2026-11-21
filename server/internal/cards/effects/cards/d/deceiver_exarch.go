package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deceiver Exarch — Creature — Phyrexian Cleric {2}{U}, 1/4 (New Phyrexia, #5575):
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 When this creature enters, choose one —
//	 • Untap target permanent you control.
//	 • Tap target permanent an opponent controls."
//
// Flash (CR 702.8) and a modal enters trigger (Apothecary Stomper's
// shape): the bullet is chosen as the trigger goes on the stack (CR
// 603.3c), and each bullet targets a permanent (CR 603.3d).
//
// No simplification.
func init() {
	trigger := effects.WhenThisEnters("Deceiver Exarch — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseOne(
		effects.ModeDoing("Untap target permanent you control.",
			effects.TargetPermanent("target permanent you control", effects.YouControl()), effects.UntapTheModesTarget),
		effects.ModeDoing("Tap target permanent an opponent controls.",
			effects.TargetPermanent("target permanent an opponent controls", effects.OpponentControls()), effects.TapTheModesTarget),
	)
	effects.Register(effects.Spec{
		OracleID:        "3c939ea6-68b7-4965-b1d3-af1d3dc79778",
		Name:            "Deceiver Exarch",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
