package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// ManaAbilitiesOfThisPermanentCostLife is "Mana abilities of this land
// cost an additional N life to activate" (Thran Portal, #5661): every
// mana ability the permanent has, printed, intrinsic to a land type it
// is (CR 305.6) or granted by another effect, costs `life` more life
// (CR 118.8, CR 601.2f through CR 602.2b). The engine adds it to each
// row's life cost (game.Characteristic.ManaAbilityLifeTax), so the
// activation, the cost chip, the auto-tapper and a bot all read one
// number. It stops with the permanent's other abilities (CR 305.7).
//
//nolint:unparam // the printed amount, kept at the call site
func ManaAbilitiesOfThisPermanentCostLife(life int) game.StaticAbility {
	return game.StaticAbility{
		Label:     "Mana abilities of this permanent cost additional life to activate.",
		Layer:     game.Layer6Ability,
		AppliesTo: thisPermanent,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.ManaAbilityLifeTax += life
		},
	}
}
