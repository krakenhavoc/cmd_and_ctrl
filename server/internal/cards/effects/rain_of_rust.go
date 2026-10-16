package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rain of Rust — Instant {3}{R}{R}:
//
//	"Choose one —
//	 • Destroy target artifact.
//	 • Destroy target land.
//	 Entwine {3}{R} (Choose both if you pay the entwine cost.)"
//
// Two destroy bullets, each with its own target; entwined (CR 702.42,
// #3866) both, the artifact first (CR 702.42b). An artifact land may be
// the target of both (one target word per bullet, CR 115.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "074e467e-f574-4b44-b5a3-798b592a5a87",
		Name:          "Rain of Rust",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{3}{R}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Destroy target artifact.", TargetPermanent("target artifact", Artifact()), DestroyTheModesTarget),
			ModeDoing("Destroy target land.", TargetPermanent("target land", Land()), DestroyTheModesTarget),
		)),
	})
}
