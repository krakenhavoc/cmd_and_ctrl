package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ashaya's Enduring Bond — Legendary Sorcery {1}{R}{G} (Mystery Booster
// Commander, #5948):
//
//	"(You may cast a legendary sorcery only if you control a legendary
//	 creature or planeswalker.)
//	 Discover X, where X is the amount of mana spent to cast this spell.
//	 Ashaya's Enduring Bond can be your commander."
//
// The cast gate is the legendary-sorcery one (CR 205.4e). X is every mana spent
// to cast the spell, read from its own payment record as it resolves (Hurl into
// History's discover with Archaic's Agony's reading); a payment the engine
// waived leaves X at 0, which is weaker, never stronger. Discover is ADR 0099's.
// "Can be your commander" is deck-building and is read from the oracle text by
// the deck validator, not here.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "0339d25b-9591-49c0-bf31-4ba749ebd0b0",
		Name:               "Ashaya's Enduring Bond",
		Completeness:       effects.CompletenessFull,
		Discovers:          true,
		ReadsManaSpent:     game.ReadsOwnPayment,
		CastCondition:      effects.LegendarySorcery(),
		CastConditionLabel: effects.LegendarySorceryLabel,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.Discover{N: ctx.ManaSpent().Total()}.Apply(ctx)
		},
	})
}
