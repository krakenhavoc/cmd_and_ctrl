package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lava Dart — Instant {R}:
//
//	"Lava Dart deals 1 damage to any target.
//	 Flashback—Sacrifice a Mountain. (You may cast this card from your
//	 graveyard for its flashback cost. Then exile it.)"
//
// A #1727 proof card: the second flashback whose whole price is a
// sacrifice, and the one whose fodder is a LAND rather than a
// creature — so the clause's predicate, not a creature assumption in
// the engine, decides what may pay. HasSubtype reads effective
// subtypes, so a land turned into a Mountain counts.
func init() {
	Register(Spec{
		OracleID:      "e48891e3-30a2-4fc8-a858-cec33c6e4ab5",
		Name:          "Lava Dart",
		Completeness:  CompletenessFull,
		Targets:       TargetAny(),
		Purpose:       ForTargets(DamageToTarget(0, 1)),
		CastableZones: []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{
			FlashbackSacrifice(1, "a Mountain", HasSubtype("Mountain")),
		},
		OnResolve: damageToFirstTarget(1),
	})
}
