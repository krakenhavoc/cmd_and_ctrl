package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Undergrowth — Instant {G}:
//
//	"As an additional cost to cast this spell, you may pay {2}{R}.
//	 Prevent all combat damage that would be dealt this turn. If this
//	 spell's additional cost was paid, this effect doesn't affect combat
//	 damage that would be dealt by red creatures."
//
// The {2}{R} is a plain optional additional cost (OptionalAdditionalMana,
// CR 118.8b), chosen as the spell is cast; it is no kicker. Unpaid, the
// spell is Fog. Paid, it is #2026's negation over a colour: red
// creatures' combat damage is dealt (the ruling), read as each creature
// would deal it (CR 609.7b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:      "f125b6b0-c5ff-44d1-bf99-ebc537d37fc3",
		Name:          "Undergrowth",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{OptionalAdditionalMana("{2}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.OptionalCostTimes(OptionalAdditionalCostKey) == 0 {
				return PreventAllCombatDamageThisTurn{Label: "Undergrowth: prevent combat damage"}.Apply(ctx)
			}
			return PreventDamageFromSource{CombatOnly: true, Protect: ShieldAnything, Filter: game.DamageSourceFilter{
				Except: []game.PermanentQuery{{Types: []string{"creature"}, Colors: []string{"R"}}},
			}}.Apply(ctx)
		},
	})
}
