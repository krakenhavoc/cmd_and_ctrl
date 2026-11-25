package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Make Mischief — Sorcery {2}{R} (Eldritch Moon, S82 OLD, #5541):
//
//	"Make Mischief deals 1 damage to any target. Create a 1/1 red Devil
//	 creature token. It has "When this token dies, it deals 1 damage to any
//	 target.""
//
// The damage to the target, then the Devil (the token is created whether
// or not the target is still legal, CR 608.2b: only the damage needs it).
// The Devil's death trigger is the token's own and targets as it dies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5de50aa5-3b30-4d7d-9203-bc8fc0f8fb34",
		Name:         "Make Mischief",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      makeMischiefPurpose(),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.SourceDealsDamageToEachLegalTarget(1)(ctx.Game, ctx.Item)
			}),
			effects.CreateToken{Template: effects.S82OldB276DevilToken(), N: 1},
		),
	})
}

func makeMischiefPurpose() game.Purpose {
	p := effects.ForTargets(effects.DamageToTarget(0, 1))
	p.Tokens, p.CreatureTokens = 1, 1
	return p
}
