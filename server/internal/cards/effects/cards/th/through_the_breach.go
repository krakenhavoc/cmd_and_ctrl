package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "c207aaee-0b67-4520-ad02-2c289228be2a",
		Name:         "Through the Breach",
		Purpose:      game.Purpose{Answers: game.AnswerMakesBlocker},
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoArcane("{2}{R}{R}"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PutFromHandOntoBattlefield{
				Match:    effects.Creature(),
				Optional: true,
				Label:    "Through the Breach — you may put a creature card from your hand onto the battlefield",
				ThenTail: throughTheBreachEntered,
			}.Apply(ctx)
		},
	})
}

var throughTheBreachEntered = effects.PutFromHandTail("through-the-breach/entered", func(ctx *effects.Context, res effects.PutFromHandResult) error {
	if res.Entered == uuid.Nil {
		return nil
	}
	c, ok := effects.BattlefieldCardForEffect(ctx.Game, res.Entered)
	if !ok {
		return nil
	}
	if err := (effects.GrantKeywordUntilEOT{Target: res.Entered, Keywords: []string{"haste"},
		Label: "Through the Breach — haste"}).Apply(ctx); err != nil {
		return err
	}
	return effects.ScheduleDelayedTrigger{
		Label:  "Through the Breach — sacrifice that creature",
		Body:   effects.StdB66MeekAttackSacrifice,
		Params: game.EffectParams{Object: game.ObjectRef{ID: res.Entered, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
})
