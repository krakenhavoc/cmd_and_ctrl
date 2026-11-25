package effects

import (
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// most_common_color.go — #6207: "the most common color among all
// permanents or tied for most common". The count is
// game.MostCommonColorsForEffect (layered colours, CR 105.2); these are
// the shapes the cards print.

// ColorIsMostCommon reports whether `color` (an uppercase WUBRG letter)
// is the most common colour among all permanents or is tied for most
// common.
func ColorIsMostCommon(g *game.Game, color string) bool {
	return slices.Contains(g.MostCommonColorsForEffect(uuid.Nil, false), strings.ToUpper(color))
}

// ColorIsMostCommonUntiedAmongNontoken reports whether `color` is the
// most common colour among nontoken permanents `player` controls and
// isn't tied for most common (Call to Arms).
func ColorIsMostCommonUntiedAmongNontoken(g *game.Game, player uuid.UUID, color string) bool {
	most := g.MostCommonColorsForEffect(player, true)
	return len(most) == 1 && most[0] == strings.ToUpper(color)
}

// SharesAMostCommonColor reports whether `c` shares a colour with the
// most common colour among all permanents or a colour tied for most
// common (Heroic Defiance, Tsabo's Assassin, Barrin's Unmaking). Read
// through the layered view; a colourless object shares no colour.
func SharesAMostCommonColor(g *game.Game, c *game.Card) bool {
	if c == nil {
		return false
	}
	most := g.MostCommonColorsForEffect(uuid.Nil, false)
	for _, col := range c.EffectiveColors() {
		if slices.Contains(most, strings.ToUpper(col)) {
			return true
		}
	}
	return false
}

// SelfPumpWhileColorIsMostCommon is "This creature gets +p/+t as long as
// <color> is the most common color among all permanents or is tied for
// most common" (the Mirage Djinns' -2/-2): a layer 7c modification
// (CR 613.4c) whose condition reads colours, which layer 5 settled for
// the pass already.
func SelfPumpWhileColorIsMostCommon(color string, power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label: "This creature gets " + SignedPT(power, toughness) + " as long as " + colorWords([]string{color}) +
			" is the most common color among all permanents or is tied for most common.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return SelfCreature(target, g, source) && ColorIsMostCommon(g, color)
		},
		Apply: PumpBy(power, toughness),
	}
}

// PumpEnchantedUnlessSharesAMostCommonColor is "Enchanted creature gets
// +p/+t unless it shares a color with the most common color among all
// permanents or a color tied for most common" (Heroic Defiance).
func PumpEnchantedUnlessSharesAMostCommonColor(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label: "Enchanted creature gets " + SignedPT(power, toughness) +
			" unless it shares a color with the most common color among all permanents or a color tied for most common.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return AttachedToSource(target, g, source) && target.IsCreature() && !SharesAMostCommonColor(g, target)
		},
		Apply: PumpBy(power, toughness),
	}
}
