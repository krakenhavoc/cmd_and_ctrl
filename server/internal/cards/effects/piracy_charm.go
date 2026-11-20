package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "dffde884-15bc-4985-90e2-a0888ef4cb48",
		Name:         "Piracy Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Target creature gains islandwalk until end of turn.",
				TargetCreature("target creature"),
				plcGrantTheModesTarget("Piracy Charm — islandwalk until end of turn", "islandwalk")),
				ForTargets(BenefitUntilEOT(0, 0, 0, "islandwalk"))),
			ModeDoing("Target creature gets +2/-1 until end of turn.",
				TargetCreature("target creature"),
				BoostTheModesTarget(2, -1, "Piracy Charm — +2/-1 until end of turn")),
			ModeWithPurpose(ModeDoing("Target player discards a card.", TargetPlayer("target player"), TheModesTargetDiscards(1)),
				ForTargets(game.TargetPurpose{Slot: 0, Discards: 1})),
		),
	})
}
