package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunder Magic — Instant {R} (#3777):
//
//	"Tiered (Choose one additional cost.)
//	 • Thunder — {0} — Thunder Magic deals 2 damage to target creature.
//	 • Thundara — {3} — Thunder Magic deals 4 damage to target creature.
//	 • Thundaga — {5}{R} — Thunder Magic deals 8 damage to target creature."
//
// Tiered proof card (CR 702.183a): three bullets that differ only in
// price and amount, so the price is what the test pins.
//
// No simplification.
func init() {
	bullet := func(label, cost string, n int) game.ModeOption {
		return effects.ModeWithPurpose(effects.SpreeModeDoing(label, cost, effects.TargetCreature("target creature"),
			func(item *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
			}), effects.ForTargets(effects.DamageToTarget(0, n)))
	}
	effects.Register(effects.Spec{
		OracleID:     "623c1873-cd83-4a5f-b391-fd0d2bb89061",
		Name:         "Thunder Magic",
		Completeness: effects.CompletenessFull,
		Modes: effects.Tiered(
			bullet("Thunder — Thunder Magic deals 2 damage to target creature.", "{0}", 2),
			bullet("Thundara — Thunder Magic deals 4 damage to target creature.", "{3}", 4),
			bullet("Thundaga — Thunder Magic deals 8 damage to target creature.", "{5}{R}", 8),
		),
	})
}
