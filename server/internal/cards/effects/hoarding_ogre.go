package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hoarding Ogre — "Whenever this creature attacks, roll a d20. 1–9: Create
// a Treasure token. 10–19: Create two Treasure tokens. 20: Create three
// Treasure tokens."
func init() {
	Register(Spec{
		OracleID:     "faeee3a7-c66c-48d8-a994-12377d3bb729",
		Name:         "Hoarding Ogre",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Hoarding Ogre — roll a d20 and create Treasures", Do(RollDice{Sides: 20, N: 1, Then: hoardingOgreTreasures})),
		},
	})
}

// hoardingOgreTreasures is the results table, after the d20 (#4307).
var hoardingOgreTreasures = DiceTail("dice/hoarding-ogre-treasures", func(ctx *Context, rolls []int) error {
	n := 1
	if rolls[0] >= 10 {
		n = 2
	}
	if rolls[0] == 20 {
		n = 3
	}
	return treasureTokens(ctx.Game, ctx.Controller(), n)
})
