package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bull-Rush Bruiser — Creature — Minotaur Warrior {3}{R}, 4/3 (Battlebond,
// #5529):
//
//	"Whenever this creature attacks, if your team controls another
//	 Warrior, this creature gains first strike until end of turn."
//
// "Your team" is "you" (CR 102.4). The intervening "if" is the shared
// Battlebond Warrior check (bbdWheneverAttacksWithAnotherWarrior); the
// grant is to the Bruiser itself, and a Bruiser that left and came back is
// a new object and gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d5a7412a-2258-4d3e-9e0f-a150cc69081e",
		Name:         "Bull-Rush Bruiser",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			bbdWheneverAttacksWithAnotherWarrior("Bull-Rush Bruiser — gains first strike until end of turn",
				thisGainsKeywordUntilEndOfTurn("first strike", "Bull-Rush Bruiser — first strike until end of turn")),
		},
	})
}
