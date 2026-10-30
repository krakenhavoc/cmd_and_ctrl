package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lurking Lizards — Creature — Lizard Villain {1}{G}, 1/3 (Marvel Super
// Heroes Commander, #4455):
//
//	"Trample
//	 Whenever you cast a spell with mana value 4 or greater, put a +1/+1
//	 counter on this creature."
//
// The cast trigger is Superior Foes of Spider-Man's WheneverYouCast with
// ManaValueGE(4); the counter lands on the Lizards if they are still the
// same permanent (CounterOnThis).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "616312a9-83c8-4fb7-87ce-079ccd3231fa",
		Name:            "Lurking Lizards",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(ManaValueGE(4), "Lurking Lizards — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
