package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rending Volley — Instant {R} (Avatar: The Last Airbender Eternal, #5075):
//
//	"This spell can't be countered.
//	 Rending Volley deals 4 damage to target white or blue creature."
//
// Combust's shape without its "can't be prevented": the spell's own
// CantBeCountered flag and 4 damage to a creature that is white or blue
// as it is targeted and again as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0324941f-484c-4656-8d1c-66e919c2bc7c",
		Name:            "Rending Volley",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		Targets:         effects.TargetCreature("target white or blue creature", effects.Or(effects.OfColor("W"), effects.OfColor("U"))),
		Purpose:         effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve:       effects.DamageToFirstTargetStep(4),
	})
}
