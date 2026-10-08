package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sheer Drop — Sorcery {2}{W}:
//
//	"Destroy target tapped creature.
//	 Awaken 3—{5}{W}"
//
// ADR 0135 §3 (#2411): Royal Assassin's "target tapped creature", then the
// awaken land (CR 702.113a), each target checked on its own (CR 608.2b): a
// creature that untapped in response is no longer a legal target.
//
// No simplifications.
func init() {
	t := TargetCreature("target tapped creature", tappedPermanent())
	Register(Spec{
		OracleID:     "7c42123d-c8e2-4571-b07d-f782e9ad1f8b",
		Name:         "Sheer Drop",
		Completeness: CompletenessFull,
		Targets:      t,
		AlternativeCosts: []game.AlternativeCost{
			Awaken(3, "{5}{W}", t),
		},
		OnResolve: AwakenAfter(3, destroyAwakenSpellTarget),
	})
}
