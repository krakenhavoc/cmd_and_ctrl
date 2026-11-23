package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karai's Technique — Sorcery {1}{W}{B} (TMT, #4457):
//
//	"Sneak {W}{B} (You may cast this spell for {W}{B} if you also return
//	 an unblocked attacker you control to hand during the declare
//	 blockers step.)
//	 Choose one or both —
//	 • Target creature gets +3/+3 until end of turn.
//	 • Target creature gets -3/-3 until end of turn."
//
// Sneak is Sneak(cost) (CR 702.190). The modes are Borrowed
// Malevolence's, without escalate and at three: each bullet pumps its
// own target while that target is still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "e53fb128-bf90-4af6-b998-ef9c250001cb",
		Name:             "Karai's Technique",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Sneak("{W}{B}")},
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets +3/+3 until end of turn.",
				effects.TargetCreature("target creature"),
				effects.BoostTheModesTarget(3, 3, "Karai's Technique")), effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3))),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets -3/-3 until end of turn.",
				effects.TargetCreature("target creature"),
				effects.BoostTheModesTarget(-3, -3, "Karai's Technique")), effects.ForTargets(effects.HarmUntilEOT(0, -3, -3))),
		),
	})
}
