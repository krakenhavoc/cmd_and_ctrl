package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// lose_abilities_except_mana.go — #6246: "All lands lose all abilities
// except mana abilities." (Blood Sun). The engine half is
// game/lose_abilities_except_mana.go.
//
//	Static: []game.StaticAbility{LandsLoseAllAbilitiesExceptMana()}

// LandsLoseAllAbilitiesExceptMana is Blood Sun's static: every land on
// the battlefield loses all its abilities except its mana abilities (CR
// 613.1f, CR 605.1a). Its triggered, non-mana activated, static and
// replacement abilities and its keywords are gone; its own mana
// abilities, the mana ability of its basic land types and a granted
// mana ability stay.
func LandsLoseAllAbilitiesExceptMana() game.StaticAbility {
	return game.StaticAbility{
		Label:              "All lands lose all abilities except mana abilities.",
		Layer:              game.Layer6Ability,
		RemovesAbilities:   true,
		KeepsManaAbilities: true,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsLand()
		},
		Apply: func(*game.Characteristic, *game.Card, *game.Game, *game.Card) {},
	}
}
