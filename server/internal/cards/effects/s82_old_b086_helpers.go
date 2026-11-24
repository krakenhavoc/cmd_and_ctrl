package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// AddLegendarySupertype is the layer 4 write of "is legendary" (CR 613.1d): it
// adds the Legendary supertype unless the characteristic already has it. The
// Ring's "Your Ring-bearer is legendary" and On Serra's Wings' "Enchanted
// creature is legendary" share it.
func AddLegendarySupertype(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
	for _, s := range c.Supertypes {
		if strings.EqualFold(s, "Legendary") {
			return
		}
	}
	c.Supertypes = append(c.Supertypes, "Legendary")
}

// ProtectionFromChosenColorSparingAttached is "Enchanted creature has
// protection from the chosen color. This effect doesn't remove Auras and
// Equipment you control that are already attached to it." (Benevolent
// Blessing; CR 702.16p, ProtectionFromChosenColorSparingThisAura's sibling).
// "You" is the Aura's controller as the layers run.
func ProtectionFromChosenColorSparingAttached() game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		Label:     "Enchanted creature has protection from the chosen color. This effect doesn't remove Auras and Equipment you control that are already attached to it.",
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			tok := game.ProtectionSparingAttachedOf(game.ProtectionFromColor(source.ChosenColor), source.Controller)
			c.Abilities = game.AppendKeywordAbility(c.Abilities, tok)
		},
	}
}
