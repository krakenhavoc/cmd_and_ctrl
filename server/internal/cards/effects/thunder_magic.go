package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
		return ModeWithPurpose(SpreeModeDoing(label, cost, TargetCreature("target creature"),
			func(item *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
			}), ForTargets(DamageToTarget(0, n)))
	}
	Register(Spec{
		OracleID:     "623c1873-cd83-4a5f-b391-fd0d2bb89061",
		Name:         "Thunder Magic",
		Completeness: CompletenessFull,
		Modes: Tiered(
			bullet("Thunder — Thunder Magic deals 2 damage to target creature.", "{0}", 2),
			bullet("Thundara — Thunder Magic deals 4 damage to target creature.", "{3}", 4),
			bullet("Thundaga — Thunder Magic deals 8 damage to target creature.", "{5}{R}", 8),
		),
	})
}
