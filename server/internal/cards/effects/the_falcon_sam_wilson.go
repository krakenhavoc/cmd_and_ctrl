package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Falcon, Sam Wilson — {4}{W} Legendary Creature — Human Hero 3/3
// (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 When The Falcon enters, put a +1/+1 counter on each other creature you
//	 control."
//
// The set of creatures is fixed before the first counter lands, and each
// counter is its own placement.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "359332d6-c42a-4dfa-b148-3296d7e74661",
		Name:            "The Falcon, Sam Wilson",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("The Falcon — put a +1/+1 counter on each other creature you control",
				mshB1CounterOnEachOtherYouControl(Creature())),
		},
	})
}
