package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flare of Malice — Instant {2}{B}{B} (Modern Horizons 3, #5948):
//
//	"You may sacrifice a nontoken black creature rather than pay this
//	 spell's mana cost.
//	 Each opponent sacrifices a creature or planeswalker with the greatest
//	 mana value among creatures and planeswalkers they control."
//
// Soul Shatter's edict (each opponent picks among the tied greatest, not
// targeted) behind Flare of Fortitude's sacrifice-instead alternative cost,
// paid as the spell goes on the stack (CR 118.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ce156f8-6b0b-41ce-9e8f-b2735c757406",
		Name:         "Flare of Malice",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.SacrificeInstead(1, "a nontoken black creature", 0, effects.Creature(), effects.OfColor("B"), effects.Not(effects.IsTokenPredicate())),
		},
		// Soul Shatter's body, looked up when the spell resolves (a file that
		// registers earlier cannot read the registry at init).
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			soulShatter, ok := effects.Lookup(soulShatterOracleID)
			if !ok || soulShatter.OnResolve == nil {
				return nil
			}
			return soulShatter.OnResolve(item, ctx)
		},
	})
}

// soulShatterOracleID is Soul Shatter's catalog row, whose edict Flare of
// Malice shares.
const soulShatterOracleID = "615927c2-3fb0-4e64-a1a7-55fb56de1423"
