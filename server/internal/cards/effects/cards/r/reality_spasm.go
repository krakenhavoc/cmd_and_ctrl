package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reality Spasm — Instant {X}{U}{U} (Rise of the Eldrazi, S82 OLD b485):
//
//	"Choose one —
//	 • Tap X target permanents.
//	 • Untap X target permanents."
//
// Each bullet names exactly X targets, X being the number announced with
// the cast (CountFromX, the Silkguard shape), so X = 0 names none and does
// nothing. Every target still legal as it resolves is tapped or untapped
// (CR 608.2b); one that left is skipped.
//
// No simplification.
func init() {
	tapTargets := effects.TargetPermanent("X target permanents")
	tapTargets.CountFromX = true
	untapTargets := effects.TargetPermanent("X target permanents")
	untapTargets.CountFromX = true
	effects.Register(effects.Spec{
		OracleID:     "34a2fc4c-ab56-4ecd-95d9-04316046843c",
		Name:         "Reality Spasm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Tap X target permanents.", tapTargets, effects.S82OldB485TapModeTargets),
			effects.ModeDoing("Untap X target permanents.", untapTargets, effects.S82OldB485UntapModeTargets),
		),
	})
}
