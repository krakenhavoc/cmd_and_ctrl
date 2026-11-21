package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ramses Overdark — Legendary Creature — Human Assassin {2}{U}{U}{B}{B},
// 4/3 (Legends, #5639):
//
//	"{T}: Destroy target enchanted creature."
//
// The target clause is a creature with an Aura attached (Enchanted, a
// Role included, CR 303.4), any controller's, read at announce and
// again as it resolves, so a creature whose Aura left in response is no
// longer a legal target (CR 608.2b). The body is the shared
// destroyFirstLegalTarget, as Royal Assassin's. The tap means summoning
// sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c9a9071-8159-4607-9465-de796f1dc2cc",
		Name:         "Ramses Overdark",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Destroy target enchanted creature.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target enchanted creature", effects.Enchanted()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
