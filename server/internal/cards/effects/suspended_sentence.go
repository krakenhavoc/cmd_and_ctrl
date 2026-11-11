package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Suspended Sentence — Instant {3}{B} (Duskmourn Commander):
//
//	"Destroy target creature an opponent controls. That player loses 3
//	 life. Exile Suspended Sentence with three time counters on it.
//	 Suspend 3—{1}{B} (...)"
//
// The controller is read before the destroy, so the life loss lands on the
// player who controlled the creature even when it was indestructible and
// stayed. The spell then exiles itself with three time counters
// (cmdr2S40ExileWithTimeCounters), as Rousing Refrain does; a spell whose
// target became illegal does not resolve, goes to the graveyard and is not
// exiled, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "98641070-93bd-4088-90cb-956c2b21979e",
		Name:           "Suspended Sentence",
		Completeness:   CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		SpecialActions: []game.SpecialAction{Suspend(3, "{1}{B}")},
		Targets:        TargetCreature("target creature an opponent controls", OpponentControls()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			controller, ok := controllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, controller, -3); err != nil {
				return err
			}
			return cmdr2S40ExileWithTimeCounters(ctx, 3)
		},
	})
}
