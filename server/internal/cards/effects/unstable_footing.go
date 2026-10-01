package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unstable Footing — Instant {R}:
//
//	"Kicker {3}{R}
//	 Damage can't be prevented this turn. If this spell was kicked, it
//	 deals 5 damage to target player or planeswalker."
//
// The target exists only when the spell is kicked (WhenPaid on the
// kicker, #1716): an unkicked Unstable Footing targets nothing and is
// just the turn grant (ADR 0107 §5). A kicked one whose target becomes
// illegal is countered on resolution (CR 608.2b), grant and all.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "fb901095-2ea4-4b9e-ad5a-2892894dc4c3",
		Name:         "Unstable Footing",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{
			WhenPaid(Kicker("{3}{R}"), targetPlayerOrPlaneswalker()),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DamageCantBePreventedThisTurn{}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.WasKicked() {
				return nil
			}
			return damageToFirstTarget(5)(item, ctx)
		},
	})
}
