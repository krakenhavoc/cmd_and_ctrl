package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twisted Reflection — Instant {1}{U}:
//
//	"Choose one —
//	 • Target creature gets -6/-0 until end of turn.
//	 • Switch target creature's power and toughness until end of turn.
//	 Entwine {B} (Choose both if you pay the entwine cost.)"
//
// The switch is layer 7d (CR 613.4d, #3996), after every set, pump and
// counter, so entwined on one creature the -6/-0 (7c) applies first and
// the switch second whatever order they resolve in: a 6/6 becomes 0/6
// and then 6/0, and dies. Each bullet has its own target (CR 702.42).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "57fdf496-c31e-4a3b-a8a6-4f053b808267",
		Name:          "Twisted Reflection",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Entwine("{B}")},
		Modes: effects.Entwining(effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets -6/-0 until end of turn.", effects.TargetCreature("target creature"),
				effects.BoostTheModesTarget(-6, 0, "Twisted Reflection — -6/-0")), effects.ForTargets(effects.HarmUntilEOT(0, -6, 0))),
			effects.ModeDoing("Switch target creature's power and toughness until end of turn.", effects.TargetCreature("target creature"),
				effects.SwitchTheModesTarget("Twisted Reflection — switch power and toughness")),
		)),
	})
}
