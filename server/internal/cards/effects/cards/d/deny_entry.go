package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deny Entry — Instant {2}{U} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"Counter target creature spell. Draw a card, then discard a card."
//
// Essence Scatter's target with Refute's body (counterTheTargetThenLootOne):
// a target that left the stack in response leaves Deny Entry with no legal
// target, so it does not resolve and nobody loots (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7dad8fe-b0b8-4c6d-ad11-76b402952799",
		Name:         "Deny Entry",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target creature spell", effects.Creature()),
		OnResolve:    effects.CounterTheTargetThenLootOne,
	})
}
