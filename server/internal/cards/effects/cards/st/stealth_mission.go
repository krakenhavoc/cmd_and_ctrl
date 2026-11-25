package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stealth Mission — Sorcery {2}{U} (War of the Spark):
//
//	"Put two +1/+1 counters on target creature you control. That creature can't
//	 be blocked this turn."
//
// Both sentences are about the one target, in order: the counters, then the
// evasion until end of turn (CR 509.1b), each only while the creature is still a
// legal target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92419fbf-7a1b-428f-92d6-e9fd510dbfac",
		Name:         "Stealth Mission",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
					return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.TargetCantBeBlockedThisTurn("Stealth Mission")(ctx.Game, ctx.Item)
			}),
		),
	})
}
