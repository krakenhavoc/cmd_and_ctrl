package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urza's Rage — Instant {2}{R}:
//
//	"Kicker {8}{R}
//	 This spell can't be countered.
//	 Urza's Rage deals 3 damage to any target. If this spell was kicked,
//	 instead it deals 10 damage to that permanent or player and the
//	 damage can't be prevented."
//
// The kicked damage's "can't be prevented" is the spell's own rider
// (SpellWasKicked, ADR 0107 §5), read off the payment record.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:                   "363f8c66-fe0c-44b9-987d-1d160e3f9c54",
		Name:                       "Urza's Rage",
		Completeness:               CompletenessFull,
		CantBeCountered:            true,
		SpellDamageCantBePrevented: SpellWasKicked(),
		OptionalCosts:              []game.AdditionalCost{Kicker("{8}{R}")},
		Targets:                    TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			amount := 3
			if ctx.WasKicked() {
				amount = 10
			}
			return damageToFirstTarget(amount)(item, ctx)
		},
	})
}
