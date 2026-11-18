package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Djinni Windseer — Creature — Djinn {3}{U}, 3/3 (Adventures in the
// Forgotten Realms, #5499):
//
//	"Flying
//	 When this creature enters, roll a d20.
//	 1—9 | Scry 1.
//	 10—19 | Scry 2.
//	 20 | Scry 3."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f06cecd4-8447-4351-9dfa-861afb84f1e6",
		Name:            "Djinni Windseer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Djinni Windseer — roll a d20 and scry", Do(RollDice{Sides: 20, N: 1, Then: djinniWindseerScry})),
		},
	})
}

// djinniWindseerScry is the results table, after the d20: scry 1, 2 or 3.
var djinniWindseerScry = DiceTail("dice/djinni-windseer-scry", func(ctx *Context, rolls []int) error {
	return Scry{N: afrB10Count(rolls[0])}.Apply(ctx)
})
