package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Canyon Crab — Creature — Crab {1}{U}, 0/5 (Outlaws of Thunder
// Junction, #4468):
//
//	"{1}{U}: This creature gets +2/-2 until end of turn.
//	 At the beginning of your end step, if you haven't cast a spell from
//	 your hand this turn, draw a card, then discard a card."
//
// The pump is Merfolk Coralsmith's self boost, so a third activation
// kills it. The end-step loot is an intervening "if" (CR 603.4) over the
// turn's cast tally (AtYourEndStepIfNoSpellFromHand, #5235): checked as
// the end step begins and again as it resolves, so a spell cast from the
// hand in response turns it off, and one cast from exile (a plotted
// card) or a graveyard does not. The discard is chosen after the draw
// (b16DrawThenDiscard).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4681facf-21ca-4118-a827-0e3d4233d8e6",
		Name:         "Canyon Crab",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}: This creature gets +2/-2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{U}"),
			Effect:  thisGetsUntilEndOfTurn(2, -2, "Canyon Crab — +2/-2 until end of turn"),
		}},
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfNoSpellFromHand("Canyon Crab — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return b16DrawThenDiscard(g, item, 1) }),
		},
	})
}
