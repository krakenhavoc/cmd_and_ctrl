package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iron Verdict — Instant {2}{W} (Kaldheim, #5505):
//
//	"Iron Verdict deals 5 damage to target tapped creature.
//	 Foretell {W}"
//
// Aerial Assault's "target tapped creature" (tappedPermanent, re-checked
// as the spell resolves, so a creature untapped in response is no
// longer a legal target) with 5 damage from the spell
// (damageToFirstTarget). Foretell is the shared special action
// (CR 702.143).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "70528ca9-195c-49a5-9d83-d51df3b0591f",
		Name:           "Iron Verdict",
		Completeness:   CompletenessFull,
		Targets:        TargetCreature("target tapped creature", tappedPermanent()),
		Purpose:        ForTargets(DamageToTarget(0, 5)),
		SpecialActions: []game.SpecialAction{Foretell("{W}")},
		OnResolve:      damageToFirstTarget(5),
	})
}
