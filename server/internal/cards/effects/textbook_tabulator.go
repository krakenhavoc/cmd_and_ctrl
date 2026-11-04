package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Textbook Tabulator — Creature — Frog Wizard {2}{U}, 0/3 (Secrets of
// Strixhaven, #4456):
//
//	"Increment (Whenever you cast a spell, if the amount of mana you spent
//	 is greater than this creature's power or toughness, put a +1/+1
//	 counter on this creature.)
//	 When this creature enters, surveil 2."
//
// Increment is the keyword helper (CR 702.191), as Hungry Graffalon; the
// enters trigger is a plain surveil 2 (CR 701.25).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ebb40a50-cfbc-4ec1-9c54-26a1bc2e7a6e",
		Name:            "Textbook Tabulator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{KeywordIncrement},
		ReadsManaSpent:  game.ReadsYourCasts,
		Triggered: []game.TriggeredAbility{
			Increment(),
			WhenThisEnters("Textbook Tabulator — surveil 2", Do(Surveil{N: 2})),
		},
	})
}
