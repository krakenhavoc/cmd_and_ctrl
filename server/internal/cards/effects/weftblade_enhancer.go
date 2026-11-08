package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Weftblade Enhancer — Creature — Drix Artificer {5}{W}, 3/4 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, put a +1/+1 counter on each of up to two
//	 target creatures.
//	 Warp {2}{W}"
//
// Rishkar, Peema Renegade's enters trigger
// (putPlusOneCounterOnEachLegalTarget) with Knight Luminary's warp offer.
// Any creatures may be chosen, the Enhancer itself included; a pick that
// is no longer legal as the trigger resolves gets nothing (CR 608.2b).
// Counters put on other creatures stay when a warped Enhancer is exiled at
// the next end step (CR 702.185a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "4d8ee903-17b3-43c9-9637-ee596955a9f4",
		Name:             "Weftblade Enhancer",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Warp("{2}{W}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Weftblade Enhancer — put a +1/+1 counter on each of up to two target creatures",
					putPlusOneCounterOnEachLegalTarget),
				TargetCreature("up to two target creatures").WithCount(0, 2)),
		},
	})
}
