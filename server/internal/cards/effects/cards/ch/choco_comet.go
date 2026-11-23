package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Choco-Comet — Sorcery {X}{R}{R} (Final Fantasy, #4462):
//
//	"Choco-Comet deals X damage to any target.
//	 Create a 2/2 green Bird creature token with "Whenever a land you
//	 control enters, this token gets +1/+0 until end of turn.""
//
// Blaze's X damage (dealXToTarget, X read as the spell resolves), then
// the set's Chocobo Bird (chocoboBirdToken). The two sentences are two
// steps (CR 608.2c), so the Bird waits for any question the damage asks.
// A target gone by resolution counters the spell (CR 608.2b): no damage
// and no Bird. X = 0 deals no damage and still makes the Bird.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "07b28ed7-eab8-4560-9994-3befcba876d3",
		Name:         "Choco-Comet",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageXToTarget(0)), Tokens: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DealXToTarget(ctx.Item, ctx) }),
			effects.CreateToken{Template: effects.ChocoboBirdToken(), N: 1},
		),
	})
}
