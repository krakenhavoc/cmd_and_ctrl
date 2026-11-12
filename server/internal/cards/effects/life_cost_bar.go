package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// PlayersCantPayLifeToCastOrActivate is "Players can't pay life to cast
// spells or to activate abilities that aren't mana abilities." (Karn's
// Sylex, #5660): the game.PlayersCantPayLifeToCastOrActivate world rule,
// carried on the source itself. While the permanent is on the
// battlefield with its abilities, game.CanPayLifeToCastOrActivateLocked
// refuses every life component of a spell's cost and of a non-mana
// activation's (game/life_cost_bar.go); a mana ability's cost and every
// other payment are untouched.
func PlayersCantPayLifeToCastOrActivate() game.StaticAbility {
	return game.StaticAbility{
		Label:     "Players can't pay life to cast spells or to activate abilities that aren't mana abilities.",
		Layer:     game.Layer6Ability,
		AppliesTo: selfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Restrictions |= game.PlayersCantPayLifeToCastOrActivate
		},
	}
}
