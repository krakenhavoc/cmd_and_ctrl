package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demon Bolt — Instant {2}{R} (Kaldheim, #5505):
//
//	"Demon Bolt deals 4 damage to target creature or planeswalker.
//	 Foretell {R}"
//
// Lightning Bolt's body (damageToFirstTarget) on Fanged Flames'
// target: 4 damage from the spell to the one target; a target gone by
// resolution counters the spell (CR 608.2b). Foretell is the shared
// special action (CR 702.143).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "6e786f0b-e2cb-40ee-96bc-81945a0b35a3",
		Name:           "Demon Bolt",
		Completeness:   effects.CompletenessFull,
		Targets:        effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:        effects.ForTargets(effects.DamageToTarget(0, 4)),
		SpecialActions: []game.SpecialAction{effects.Foretell("{R}")},
		OnResolve:      effects.DamageToFirstTargetStep(4),
	})
}
