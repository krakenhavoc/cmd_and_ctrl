package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancient Copper Dragon — "Whenever this creature deals combat damage to a
// player, roll a d20. You create a number of Treasure tokens equal to the
// result."
func init() {
	Register(Spec{
		OracleID:        "48daee9d-ddaf-410f-8c3a-12fa1064ab56",
		Name:            "Ancient Copper Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Ancient Copper Dragon — roll a d20 and create Treasures", Do(RollDice{Sides: 20, N: 1, Then: ancientCopperDragonTreasures})),
		},
	})
}

// ancientCopperDragonTreasures is the Treasures, after the d20 (#4307).
var ancientCopperDragonTreasures = DiceTail("dice/ancient-copper-dragon-treasures", func(ctx *Context, rolls []int) error {
	return treasureTokens(ctx.Game, ctx.Controller(), rolls[0])
})
