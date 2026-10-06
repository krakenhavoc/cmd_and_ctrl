package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ichor Explosion — Sorcery {5}{B}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 All creatures get -X/-X until end of turn, where X is the
//	 sacrificed creature's power."
//
// X is "the power of the sacrificed creature as it last existed on the
// battlefield" (the 2011-06-01 ruling, CR 608.2h), read off the payment
// record (ADR 0113 §1); a negative power is X = 0 (CR 107.1b). The
// affected set is the creatures on the battlefield as it resolves (CR
// 611.2c), as for Languish.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID: "d33aa11b-011b-4d12-85a9-4f956153fb1d",
		Name:     "Ichor Explosion",
		// ADR 0126 §6: the -X/-X is the sacrificed creature's power, counted as it is cast, so it is declared as the destruction it usually is, an upper bound.
		Purpose:        game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.SacrificedPower()
			if x == 0 {
				return nil
			}
			return BoostUntilEOT{
				Match:     Creature(),
				Power:     -x,
				Toughness: -x,
				Label:     fmt.Sprintf("Ichor Explosion — -%d/-%d", x, x),
			}.Apply(ctx)
		},
	})
}
