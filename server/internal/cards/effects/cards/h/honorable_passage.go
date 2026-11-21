package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Honorable Passage — Instant {1}{W}:
//
//	"The next time a source of your choice would deal damage to any target this turn, prevent that damage. If damage from a red source is prevented this way, Honorable Passage deals that much damage to the source's controller."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8,
// 609.7a).
// "Any target" is a target (CR 115.4). "If damage from a red source is
// prevented this way, Honorable Passage deals that much damage to the
// source's controller" is the shield's additional effect (CR 615.5); the
// colour and the controller are the source's as it would have dealt the
// damage.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66805334-1015-4b41-ba9b-116de94b0744",
		Name:         "Honorable Passage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		OnResolve:    effects.NextDamageShieldSpell(effects.PreventNextDamageFromChosenSource(effects.ShieldTheTarget).WithThen(effects.PreventedRedDamageSourceControllerBody)),
	})
}
