package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deep Wood — Instant {1}{G}:
//
//	"Cast this spell only during the declare attackers step and only if
//	 you've been attacked this step.
//	 Prevent all damage that would be dealt to you this turn by attacking
//	 creatures."
//
// Heavy Fog's text, and Heavy Fog's shape (heavy_fog.go): the cast
// restriction is a CastCondition, and the shield is #2026's combat
// status, read as each creature would deal damage (CR 609.7b), last-known
// for a creature that has left the battlefield (CR 608.2h).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "3f01f627-9fbd-470b-8001-974784ccf421",
		Name:               "Deep Wood",
		Completeness:       effects.CompletenessFull,
		CastCondition:      effects.YouWereAttackedThisStep,
		CastConditionLabel: effects.AttackedThisStepLabel,
		OnResolve:          effects.SourceShieldSpell(effects.DamageToYouFromAttackers()),
	})
}
