package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Razortip Whip — Artifact {2} (Gatecrash, S82 OLD #5565):
//
//	"{1}, {T}: This artifact deals 1 damage to target opponent or
//	 planeswalker."
//
// Cult Guildmage's second ability, on an artifact: a {T} cost with no
// creature sickness (CR 302.6 applies to creatures only).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "44cf7fd5-721f-41f4-ae78-cf93afeff5f1",
		Name:         "Razortip Whip",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: This artifact deals 1 damage to target opponent or planeswalker.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetOpponentOrPlaneswalker(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
