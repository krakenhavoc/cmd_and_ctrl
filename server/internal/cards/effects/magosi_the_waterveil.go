package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magosi, the Waterveil — Land:
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 {U}, {T}: Put an eon counter on this land. Skip your next turn.
//	 {T}, Remove an eon counter from this land and return it to its
//	 owner's hand: Take an extra turn after this one."
//
// Four lines, each an existing shape plus the one this card was waiting
// for. The tapped entry is the CR 614 self-replacement and the blue mana
// a plain tap ability. The third line is a CR 602 ability that puts the
// counter and then skips: SkipYourNextTurns writes a skip debt (CR 614.10,
// ADR 0188 Decision 1) that the rotation seam spends when the controller's
// next turn would begin, so the turn in progress is never cut short. The
// skip does not depend on the counter, as the 2009-10-01 ruling says
// ("you'll skip your next turn even if you don't put an eon counter on
// Magosi"), and it does not depend on Magosi still being there when the
// ability resolves. The fourth line pays its counter and the return to
// hand as costs at announce, and queues the extra turn when it resolves.
//
// The same ruling covers using both lines in one turn: they cancel out,
// because the next turn the player would begin is the extra turn the
// fourth line made, and that is the one the skip is spent on. The engine
// gets that for free from spending the debt at the seam.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4bdffa67-e6b3-4588-b76e-c11db6f043ca",
		Name:         "Magosi, the Waterveil",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{
			{
				Label: "{U}, {T}: Put an eon counter on this land. Skip your next turn.",
				// ADR 0188 Decision 10: the skip is a turn lost, which the bot
				// prices (negative) against the eon counter.
				Purpose: game.Purpose{Answers: game.AnswerValue, SkipsTurns: 1},
				Cost:    Plus(ManaCost("{U}"), TapCost()),
				Effect:  magosiPutEonCounterAndSkip,
			},
			{
				Label:   "{T}, Remove an eon counter from this land and return it to its owner's hand: Take an extra turn after this one.",
				Purpose: game.Purpose{Answers: game.AnswerValue, ExtraTurns: 1},
				Cost:    Plus(TapCost(), RemoveCountersFromThis("eon", 1), ReturnThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return TakeExtraTurn{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// magosiPutEonCounterAndSkip is the third line's effect: the counter,
// then the skip. The skip is a separate sentence, so it happens whether
// or not the counter did (a Magosi that left in response, or came back as
// a new object, gets no counter but still costs its controller the turn).
func magosiPutEonCounterAndSkip(g *game.Game, item *game.StackItem) error {
	if err := putACounterOnThis("eon")(g, item); err != nil {
		return err
	}
	return SkipYourNextTurns{}.Apply(NewContext(g, item))
}
