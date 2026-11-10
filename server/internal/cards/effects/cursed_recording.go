package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cursed Recording — Artifact {2}{R}{R} (Duskmourn: House of Horror,
// #4466):
//
//	"Whenever you cast an instant or sorcery spell, put a time counter on
//	 this artifact. Then if there are seven or more time counters on it,
//	 remove those counters and it deals 20 damage to you.
//	 {T}: When you next cast an instant or sorcery spell this turn, copy
//	 that spell. You may choose new targets for the copy."
//
// The trigger puts the counter and then reads the count as it resolves
// (the "Then if" is an instruction, not an intervening if). The tap ability
// is Loki Laufeyson's: a delayed trigger on the next matching cast this
// turn whose body is the shared copy-the-spell body (CR 707.10), the copy
// created on the stack and not cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9adc3750-0ddb-4f2c-a0e7-131e59d4cba5",
		Name:         "Cursed Recording",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()),
				"Cursed Recording — put a time counter on it; at seven, remove them and it deals 20 damage to you",
				Do(CounterOnThis{Kind: stdB26CursedRecordingCounter, N: 1}, Instruction(stdB26CursedRecordingTick))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: When you next cast an instant or sorcery spell this turn, copy that spell. You may choose new targets for the copy.",
			Cost:    TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return WhenYouNextCast("Cursed Recording — copy that spell",
					game.CastFilter{Types: []string{"Instant", "Sorcery"}}, copyTheSpellBody).Apply(NewContext(g, item))
			},
		}},
	})
}
