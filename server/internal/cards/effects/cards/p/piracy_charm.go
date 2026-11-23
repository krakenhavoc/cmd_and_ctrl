package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Piracy Charm — Instant {U} (Planar Chaos, #5595):
//
//	"Choose one —
//	 • Target creature gains islandwalk until end of turn. (It can't be blocked
//	 as long as defending player controls an Island.)
//	 • Target creature gets +2/-1 until end of turn.
//	 • Target player discards a card."
//
// Three bullets on the shared mode bodies: islandwalk until end of turn
// (CR 702.14c), a +2/-1 boost, and a discard the target player chooses
// (CR 701.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dffde884-15bc-4985-90e2-a0888ef4cb48",
		Name:         "Piracy Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gains islandwalk until end of turn.",
				effects.TargetCreature("target creature"),
				effects.PlcGrantTheModesTarget("Piracy Charm — islandwalk until end of turn", "islandwalk")),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "islandwalk"))),
			effects.ModeDoing("Target creature gets +2/-1 until end of turn.",
				effects.TargetCreature("target creature"),
				effects.BoostTheModesTarget(2, -1, "Piracy Charm — +2/-1 until end of turn")),
			effects.ModeWithPurpose(effects.ModeDoing("Target player discards a card.", effects.TargetPlayer("target player"), effects.TheModesTargetDiscards(1)),
				effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 1})),
		),
	})
}
