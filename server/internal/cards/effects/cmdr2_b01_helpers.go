package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cmdr2_b01_helpers.go — card-side helpers for the Commander-deck wave 2,
// batch 1 (simple lands) and batch 2 (mana rocks and mana triggers)
// (#5368). Every name starts with cmdr2B01.

// cmdr2B01TapForColorless is "{T}: Add {C}."
func cmdr2B01TapForColorless() ManaAbility {
	return ManaAbility{
		Cost:     ManaAbilityCost{Tap: true},
		Produced: "{C}",
		Label:    "Add {C}",
	}
}

// cmdr2B01TapForAnyColor is "{T}: Add one mana of any color."
func cmdr2B01TapForAnyColor() ManaAbility {
	return ManaAbility{
		Cost:     ManaAbilityCost{Tap: true},
		Produced: "{W|U|B|R|G}",
		Label:    "Add one mana of any color",
	}
}

// cmdr2B01TapForCreatureOnlyMana is "{T}: Add one mana of any color.
// Spend this mana only to cast a creature spell." (Ancient Ziggurat,
// Abundant Countryside; Beastcaller Savant's shape.)
func cmdr2B01TapForCreatureOnlyMana() ManaAbility {
	a := cmdr2B01TapForAnyColor()
	a.Restrictions = []string{ManaRestrictCast, ManaRestrictType("Creature")}
	a.Label = "Add one mana of any color. Spend this mana only to cast a creature spell"
	return a
}

// cmdr2B01TapForColor is a single-colour land's "{T}: Add {<c>}".
func cmdr2B01TapForColor(c string) ManaAbility {
	return ManaAbility{
		Cost:     ManaAbilityCost{Tap: true},
		Produced: "{" + c + "}",
		Label:    "Add {" + c + "}",
	}
}

// cmdr2B01MakeToken is "create a <token>" as an activated ability's whole
// effect, the token named by its tokens_table key.
func cmdr2B01MakeToken(tokenKey string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return CreateToken{
			Controller: item.Controller,
			Template:   TokenCard(tokenKey),
			N:          1,
		}.Apply(NewContext(g, item))
	}
}
