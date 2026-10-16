package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "57fdf496-c31e-4a3b-a8a6-4f053b808267",
		Name:          "Twisted Reflection",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{B}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Target creature gets -6/-0 until end of turn.", TargetCreature("target creature"),
				BoostTheModesTarget(-6, 0, "Twisted Reflection — -6/-0")),
			ModeDoing("Switch target creature's power and toughness until end of turn.", TargetCreature("target creature"),
				SwitchTheModesTarget("Twisted Reflection — switch power and toughness")),
		)),
	})
}
