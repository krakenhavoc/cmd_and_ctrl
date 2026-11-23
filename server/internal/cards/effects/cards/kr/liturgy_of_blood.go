package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "aabea4a2-5107-4abd-a0c7-ac0024d98ae0",
		Name:         "Liturgy of Blood",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DestroyTheTargetPermanent(ctx.Item, ctx) }),
			effects.AddMana{Produced: "{B}{B}{B}"},
		),
	})
}
