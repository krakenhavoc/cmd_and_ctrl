package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Through the Breach — Instant — Arcane {4}{R} (Innistrad Remastered):
//
//	"You may put a creature card from your hand onto the battlefield. That
//	 creature gains haste. Sacrifice that creature at the beginning of the
//	 next end step.
//	 Splice onto Arcane {2}{R}{R}"
//
// Sneak Attack's body for a spell: the creature enters from hand (so its
// enters-the-battlefield triggers and replacements apply), gains haste until
// end of turn and is sacrificed by a delayed trigger bound to that object
// (CR 603.7a). Splice is the engine's (CR 702.47).
func init() {
	Register(Spec{
		OracleID:     "c207aaee-0b67-4520-ad02-2c289228be2a",
		Name:         "Through the Breach",
		Purpose:      game.Purpose{Answers: game.AnswerMakesBlocker},
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{2}{R}{R}"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return PutFromHandOntoBattlefield{
				Match:    Creature(),
				Optional: true,
				Label:    "Through the Breach — you may put a creature card from your hand onto the battlefield",
				ThenTail: throughTheBreachEntered,
			}.Apply(ctx)
		},
	})
}

var throughTheBreachEntered = PutFromHandTail("through-the-breach/entered", func(ctx *Context, res PutFromHandResult) error {
	if res.Entered == uuid.Nil {
		return nil
	}
	c, ok := battlefieldCardForEffect(ctx.Game, res.Entered)
	if !ok {
		return nil
	}
	if err := (GrantKeywordUntilEOT{Target: res.Entered, Keywords: []string{"haste"},
		Label: "Through the Breach — haste"}).Apply(ctx); err != nil {
		return err
	}
	return ScheduleDelayedTrigger{
		Label:  "Through the Breach — sacrifice that creature",
		Body:   stdB66MeekAttackSacrifice,
		Params: game.EffectParams{Object: game.ObjectRef{ID: res.Entered, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
})
