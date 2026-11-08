package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "07b28ed7-eab8-4560-9994-3befcba876d3",
		Name:         "Choco-Comet",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageXToTarget(0)), Tokens: 1},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return dealXToTarget(ctx.Item, ctx) }),
			CreateToken{Template: chocoboBirdToken(), N: 1},
		),
	})
}
