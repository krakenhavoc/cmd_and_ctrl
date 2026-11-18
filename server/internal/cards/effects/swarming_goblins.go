package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Swarming Goblins — Creature — Goblin {4}{R}, 4/3 (Adventures in the
// Forgotten Realms, #5499):
//
//	"When this creature enters, roll a d20.
//	 1—9 | Create a 1/1 red Goblin creature token.
//	 10—19 | Create two of those tokens.
//	 20 | Create three of those tokens."
//
// The d20 goes through the dice door; the table is read from the result
// (CR 706.3a) and the tokens are made after the roll settles.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a038dec-2d9c-4422-a0f3-94bf70e55217",
		Name:         "Swarming Goblins",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Swarming Goblins — roll a d20 and create Goblin tokens", Do(RollDice{Sides: 20, N: 1, Then: swarmingGoblinsTokens})),
		},
	})
}

// swarmingGoblinsTokens is the results table, after the d20. An on-disk
// key: never renamed, never reused.
var swarmingGoblinsTokens = DiceTail("dice/swarming-goblins-tokens", func(ctx *Context, rolls []int) error {
	return CreateToken{Template: RedGoblinToken(), N: afrB10Count(rolls[0])}.Apply(ctx)
})
