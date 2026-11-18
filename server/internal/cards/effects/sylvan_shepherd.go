package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sylvan Shepherd — Creature — Human Druid {2}{G}, 2/3 (Adventures in the
// Forgotten Realms, #5499):
//
//	"Vigilance
//	 Whenever this creature attacks, roll a d20.
//	 1—9 | You gain 1 life.
//	 10—19 | You gain 2 life.
//	 20 | You gain 5 life."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ae9f4b85-46ed-4a1a-ada4-54ce56f41815",
		Name:            "Sylvan Shepherd",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Sylvan Shepherd — roll a d20 and gain life", Do(RollDice{Sides: 20, N: 1, Then: sylvanShepherdLife})),
		},
	})
}

// sylvanShepherdLife is the results table, after the d20.
var sylvanShepherdLife = DiceTail("dice/sylvan-shepherd-life", func(ctx *Context, rolls []int) error {
	n := 1
	if rolls[0] >= 10 {
		n = 2
	}
	if rolls[0] == 20 {
		n = 5
	}
	return GainLife{Amount: n}.Apply(ctx)
})
