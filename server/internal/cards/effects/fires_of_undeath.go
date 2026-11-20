package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fires of Undeath — Instant {2}{R} (Dark Ascension, #5571):
//
//	"Fires of Undeath deals 2 damage to any target.
//	 Flashback {5}{B} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Firebolt at instant speed with a black flashback cost (CR 702.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "205bcab6-8b6b-4d09-9890-0b7391d59788",
		Name:             "Fires of Undeath",
		Completeness:     CompletenessFull,
		Targets:          TargetAny(),
		Purpose:          ForTargets(DamageToTarget(0, 2)),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{B}")},
		OnResolve:        damageToFirstTarget(2),
	})
}
