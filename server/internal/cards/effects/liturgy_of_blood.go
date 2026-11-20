package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Liturgy of Blood — Sorcery {3}{B}{B} (Magic 2014, #5563):
//
//	"Destroy target creature. Add {B}{B}{B}."
//
// Bake into a Pie's destroy, then three black mana in its caster's pool. A
// target gone at resolution counters the spell (CR 608.2b), so no mana is
// added then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aabea4a2-5107-4abd-a0c7-ac0024d98ae0",
		Name:         "Liturgy of Blood",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return destroyTheTargetPermanent(ctx.Item, ctx) }),
			AddMana{Produced: "{B}{B}{B}"},
		),
	})
}
