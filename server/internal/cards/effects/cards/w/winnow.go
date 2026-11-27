package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Winnow — Instant {1}{W} (Invasion, S82 #5616):
//
//	"Destroy target nonland permanent if another permanent with the same name
//	 is on the battlefield.
//	 Draw a card."
//
// The name check is made as the spell resolves, against every other permanent
// on the battlefield (CR 608.2). A target that has become illegal makes the
// whole spell do nothing, including the draw (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b2ad355-3f1e-4c8b-93ee-848934646b23",
		Name:         "Winnow",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve:    WinnowResolve,
	})
}

// WinnowResolve destroys the target if another permanent shares its name,
// then draws a card.
func WinnowResolve(item *game.StackItem, ctx *effects.Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	if target, found := ctx.Game.LookupCardForEffect(t.ID); found && effects.AnotherPermanentSharesName(ctx.Game, target) {
		if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.DrawCards{N: 1}.Apply(ctx)
}
