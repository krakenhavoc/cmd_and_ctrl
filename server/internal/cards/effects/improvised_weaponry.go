package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Improvised Weaponry — Sorcery {2}{R} (Adventures in the Forgotten Realms):
//
//	"Improvised Weaponry deals 2 damage to any target. Create a Treasure
//	 token."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa5dad0b-8e23-4b88-b383-849f7318e01f",
		Name:         "Improvised Weaponry",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			if err := (DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
		},
	})
}
