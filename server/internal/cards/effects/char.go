package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Char — Instant {2}{R} (Ravnica: City of Guilds, #5600):
//
//	"Char deals 4 damage to any target and 2 damage to you."
//
// Spicy Oatmeal Pizza's body on a spell (tmtB6DamageAnyTargetAndYou): the
// 4 to the target and the 2 to its caster are dealt together, as one
// damage instance. With its only target gone the spell does not resolve
// (CR 608.2b), so the caster takes nothing either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e57d90ad-2f54-46ca-8237-aa383353a0de",
		Name:         "Char",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 4)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return tmtB6DamageAnyTargetAndYou(4, 2)(ctx.Game, item)
		},
	})
}
