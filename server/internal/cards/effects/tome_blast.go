package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tome Blast — Sorcery {1}{R} (Secrets of Strixhaven, #4456):
//
//	"Tome Blast deals 2 damage to any target.
//	 Flashback {4}{R} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// The Lightning Bolt body (damage dealt by the spell to the first target,
// re-checked at resolution, CR 608.2b) with two damage. Flashback is the
// shared constructor: the graveyard is a cast zone, the price is {4}{R}, and
// the spell is exiled instead of going anywhere else once it leaves the stack
// (CR 702.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f0163420-2f91-493c-bad4-2cd33e9856bf",
		Name:             "Tome Blast",
		Completeness:     CompletenessFull,
		Targets:          TargetAny(),
		Purpose:          game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{4}{R}")},
		OnResolve:        damageToFirstTarget(2),
	})
}
