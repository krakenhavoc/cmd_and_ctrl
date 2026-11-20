package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancient Copper Dragon — "Whenever this creature deals combat damage to a
// player, roll a d20. You create a number of Treasure tokens equal to the
// result."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "48daee9d-ddaf-410f-8c3a-12fa1064ab56",
		Name:            "Ancient Copper Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Ancient Copper Dragon — roll a d20 and create Treasures", effects.Do(effects.RollDice{Sides: 20, N: 1, Then: ancientCopperDragonTreasures})),
		},
	})
}

// ancientCopperDragonTreasures is the Treasures, after the d20 (#4307).
var ancientCopperDragonTreasures = effects.DiceTail("dice/ancient-copper-dragon-treasures", func(ctx *effects.Context, rolls []int) error {
	return effects.TreasureTokens(ctx.Game, ctx.Controller(), rolls[0])
})
