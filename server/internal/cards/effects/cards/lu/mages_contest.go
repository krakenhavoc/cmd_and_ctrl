package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mages' Contest — Instant {1}{R}{R} (Invasion, #5616):
//
//	"You and target spell's controller bid life. You start the bidding with a
//	 bid of 1. In turn order, each player may top the high bid. The bidding ends
//	 if the high bid stands. The high bidder loses life equal to the high bid. If
//	 you win the bidding, counter that spell."
//
// LifeBidding (#6212) runs the auction between the two players, opening at 1.
// The win is a registered number branch: the high bidder loses life equal to
// the high bid (CR 119.3), and when that bidder is the caster the targeted
// spell is countered, if it is still there. A bid is only a number
// (CR 107.1b), so it may be more life than the bidder has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0a6bf69-7e21-4d5c-ac72-5946d84fc2ca",
		Name:         "Mages' Contest",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: effects.ResolveSteps(effects.LifeBidding{
			Bidders: effects.S82OldB659LifeBidders,
			Opening: 1,
			Then:    magesContestWon,
		}),
	})
}

// magesContestWon is the high bidder's win. An on-disk key: never renamed,
// never reused.
var magesContestWon = effects.NumberThen("number/mages-contest-won", func(ctx *effects.Context, r game.NumberAnswer) error {
	if r.Chooser == uuid.Nil {
		return nil
	}
	if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.Chooser, -r.Number); err != nil {
		return err
	}
	if r.Chooser != ctx.Controller() {
		return nil
	}
	if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
		return effects.CounterTarget{StackID: t.ID}.Apply(ctx)
	}
	return nil
})
