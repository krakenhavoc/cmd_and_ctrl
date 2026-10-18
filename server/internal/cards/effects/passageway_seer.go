package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Passageway Seer — Creature — Tiefling Warlock {3}{B}, 2/2:
//
//	"Lifelink
//	 When this creature enters, you take the initiative.
//	 At the beginning of your end step, if you have the initiative, put
//	 a +1/+1 counter on this creature."
//
// The ETB is TakeTheInitiative (ADR 0203, #4096). The end-step clause is
// an intervening "if" (CR 603.4), IfYouHaveTheInitiative: it triggers
// only while you have the initiative and does nothing if you have lost
// it by the time it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2d8ebb48-5f37-4b7e-b8ec-3296d9a039c7",
		Name:            "Passageway Seer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouTakeTheInitiative("Passageway Seer"),
			IfYouHaveTheInitiative(AtYourEndStep("Passageway Seer — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: "+1/+1", N: 1}))),
		},
	})
}
