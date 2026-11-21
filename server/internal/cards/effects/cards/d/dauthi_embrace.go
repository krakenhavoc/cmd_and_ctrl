package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dauthi Embrace — Enchantment {2}{B} (Tempest, #5628):
//
//	"{B}{B}: Target creature gains shadow until end of turn. (It can
//	 block or be blocked by only creatures with shadow.)"
//
// Keen Glidemaster's targeted grant with shadow, through the shared
// ebTargetCreatureGainsUntilEOT body: a layer-6 grant on the target
// (CR 613.1f) that ends in the cleanup step (CR 514.2). Any creature may
// be the target, and the ability can be activated any number of times
// at instant speed. A target gone in response does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b19a7ae7-3af5-4aa6-80e1-8888054427de",
		Name:         "Dauthi Embrace",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{B}: Target creature gains shadow until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "shadow")),
			Cost:    effects.ManaCost("{B}{B}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("shadow", "Dauthi Embrace — shadow until end of turn"),
		}},
	})
}
