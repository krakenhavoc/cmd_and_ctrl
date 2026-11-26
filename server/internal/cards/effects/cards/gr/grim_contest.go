package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grim Contest — Instant {1}{B}{G} (Fate Reforged):
//
//	"Choose target creature you control and target creature an opponent controls.
//	 Each of those creatures deals damage equal to its toughness to the other."
//
// Two separate clauses; both creatures deal damage at once (CR 120.3), each
// reading its own toughness as the spell resolves, and nothing is dealt unless
// both targets are still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b78f940-f840-4d4d-a09b-147202963367",
		Name:         "Grim Contest",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok || mine.Kind != game.TargetCard {
				return nil
			}
			theirs, ok := ctx.ClauseTarget(1)
			if !ok || theirs.Kind != game.TargetCard {
				return nil
			}
			a, ok := ctx.Game.LookupCardForEffect(mine.ID)
			if !ok {
				return nil
			}
			b, ok := ctx.Game.LookupCardForEffect(theirs.ID)
			if !ok {
				return nil
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				if err := (effects.DealDamage{Source: mine.ID, Target: theirs.ID, Amount: max(a.CurrentToughness(), 0)}).Apply(ctx); err != nil {
					return err
				}
				return effects.DealDamage{Source: theirs.ID, Target: mine.ID, Amount: max(b.CurrentToughness(), 0)}.Apply(ctx)
			})
		},
	})
}
