package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cant_be_enchanted.go — #6203: "can't be enchanted" and "other players
// can't gain control of" (CR 303.4: other effects can limit what a
// permanent can be enchanted by). The engine half is
// game/cant_be_enchanted.go.
//
//	Static: []game.StaticAbility{ThisCantBeEnchanted()}                     // Tetravite: "This token can't be enchanted."
//	Static: []game.StaticAbility{EnchantedCantBeEnchantedByOtherAuras()}    // Consecrate Land, Anti-Magic Aura
//	Static: []game.StaticAbility{CantBecomeEnchantedAndOthersCantGainControl(applies, label)} // Guardian Beast

// ThisCantBeEnchanted is "This <permanent> can't be enchanted.": no Aura
// may target or become attached to it, and one attached goes to the
// graveyard (CR 303.4c).
func ThisCantBeEnchanted() game.StaticAbility {
	return game.StaticAbility{
		Label:     "This permanent can't be enchanted.",
		Layer:     game.Layer6Ability,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.ExtraRestrictions |= game.CantBeEnchanted
		},
	}
}

// EnchantedCantBeEnchantedByOtherAuras is an Aura's "Enchanted
// <permanent> … can't be enchanted by other Auras.": the Aura carries the
// rule on itself (game.EnchantedCantBeEnchantedByOthers), so whatever it
// is attached to can't be enchanted by any other Aura, and an Aura
// already there goes to the graveyard (CR 303.4c). An Aura that loses its
// abilities stops carrying it.
func EnchantedCantBeEnchantedByOtherAuras() game.StaticAbility {
	return game.StaticAbility{
		Label:     "Enchanted permanent can't be enchanted by other Auras.",
		Layer:     game.Layer6Ability,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.ExtraRestrictions |= game.EnchantedCantBeEnchantedByOthers
		},
	}
}

// CantBecomeEnchantedAndOthersCantGainControl is Guardian Beast's "<the
// permanents `applies` names> can't be enchanted … and other players
// can't gain control of them. This effect doesn't remove Auras already
// attached to those artifacts.": no new Aura may target or become
// attached to them, one already attached stays, and a gain or an exchange
// of control to any player but their controller does nothing.
// `applies` carries the condition ("as long as this creature is
// untapped") and the set ("noncreature artifacts you control").
func CantBecomeEnchantedAndOthersCantGainControl(applies func(target *game.Card, g *game.Game, source *game.Card) bool, label string) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer6Ability,
		AppliesTo: applies,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.ExtraRestrictions |= game.CantBecomeEnchanted | game.OthersCantGainControl
		},
	}
}
