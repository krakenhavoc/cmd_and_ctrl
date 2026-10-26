package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancient Gold Dragon — "Whenever this creature deals combat damage to a
// player, roll a d20. You create a number of 1/1 blue Faerie Dragon creature
// tokens with flying equal to the result."
func init() {
	Register(Spec{
		OracleID:        "44cb725a-72fc-4e0d-b966-f0b6d33f6b79",
		Name:            "Ancient Gold Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Ancient Gold Dragon — roll a d20 and create Faerie Dragons", Do(RollDice{Sides: 20, N: 1, Then: ancientGoldDragonTokens})),
		},
	})
}

// ancientGoldDragonTokens is the Faerie Dragons, after the d20 (#4307).
var ancientGoldDragonTokens = DiceTail("dice/ancient-gold-dragon-tokens", func(ctx *Context, rolls []int) error {
	t := ctx.TextToken(TokenCard("1/1 blue Faerie Dragon with flying"))
	return ctx.Game.CreateTokenForEffect(ctx.Controller(), t, rolls[0])
})
