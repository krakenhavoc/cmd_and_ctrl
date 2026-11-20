package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "f125b6b0-c5ff-44d1-bf99-ebc537d37fc3",
		Name:          "Undergrowth",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.OptionalAdditionalMana("{2}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.OptionalCostTimes(effects.OptionalAdditionalCostKey) == 0 {
				return effects.PreventAllCombatDamageThisTurn{Label: "Undergrowth: prevent combat damage"}.Apply(ctx)
			}
			return effects.PreventDamageFromSource{CombatOnly: true, Protect: effects.ShieldAnything, Filter: game.DamageSourceFilter{
				Except: []game.PermanentQuery{{Types: []string{"creature"}, Colors: []string{"R"}}},
			}}.Apply(ctx)
		},
	})
}
