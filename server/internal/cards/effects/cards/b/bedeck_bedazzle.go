package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bedeck // Bedazzle — split card {B/R}{B/R} // {4}{B}{R} (Ravnica
// Remastered):
//
//	Bedeck (Instant): "Target creature gets +3/-3 until end of turn."
//	Bedazzle (Instant): "Destroy target nonbasic land. Bedazzle deals 2
//	 damage to target opponent or planeswalker."
//
// Bedeck is Nameless Inversion's pump without the type loss. Bedazzle's two
// instructions each read their own target clause (CR 608.2b): a nonbasic
// land that has gone does not stop the damage, and vice versa.
//
// No simplification.
func bedazzlePurpose() game.Purpose {
	p := effects.ForTargets(effects.DamageToTarget(1, 2))
	p.Answers = game.AnswerRemove
	return p
}

func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8bb97a8-5f5d-48e9-a4c1-b4c20efe7739",
		Name:         "Bedeck",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, item.Targets[0].ID, nil,
				"Bedeck — +3/-3", game.ModifyPTMod(3, -3))
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "b8bb97a8-5f5d-48e9-a4c1-b4c20efe7739#1",
		Name:         "Bedazzle",
		Completeness: effects.CompletenessFull,
		Purpose:      bedazzlePurpose(),
		Targets: effects.Clauses(
			effects.TargetPermanent("target nonbasic land", effects.Land(), effects.NonbasicLand()),
			effects.TargetOpponentOrPlaneswalker(),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.KineticOozeDestroy),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(1)
				if !ok {
					return nil
				}
				return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}.Apply(ctx)
			}),
		),
	})
}
