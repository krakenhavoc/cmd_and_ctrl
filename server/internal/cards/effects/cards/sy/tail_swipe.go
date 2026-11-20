package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tail Swipe — Instant {G} (Dominaria United, #5487):
//
//	"Choose target creature you control and target creature you don't
//	 control. If you cast this spell during your main phase, the creature
//	 you control gets +1/+1 until end of turn. Then those creatures fight
//	 each other."
//
// Longstalk Brawl's two clauses with the pump conditional: whether it was
// cast in your main phase is read as the spell resolves with
// b14InYourMainPhase, the same answer as at cast time (the stack must be
// empty for a step to end). The pump goes on the creature you control only
// while it is still a legal target; if either target is illegal, neither
// creature fights (CR 608.2b, CR 701.14b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "525d98d6-b117-44f4-b19c-3dd26e7c9fd4",
		Name:         "Tail Swipe",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: a fight damages a creature
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard || !effects.B14InYourMainPhase(ctx.Game, ctx.Controller()) {
					return nil
				}
				return effects.BoostUntilEOT{Target: mine.ID, Power: 1, Toughness: 1, Label: "Tail Swipe — +1/+1"}.Apply(ctx)
			}),
			effects.Instruction(effects.FightTargetClauses),
		),
	})
}
