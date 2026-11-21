package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cryogenic Stasis — Instant {1}{U} (Star Trek Commander, #4974):
//
//	"Tap target creature and put a stun counter on it.
//	 Draw a card."
//
// Cryotheory Adept's tap-and-stun on the spell's target (the stun counter
// is CR 122.1d's untap replacement, which the engine applies), then the
// draw. A target gone by resolution counters the spell (CR 608.2b), so
// nothing is drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "846ba66c-a4eb-4d00-a0f1-d9f9524b9695",
		Name:         "Cryogenic Stasis",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.TapTargetAndStunIt(ctx.Game, ctx.Item) }),
			effects.DrawCards{N: 1},
		),
	})
}
