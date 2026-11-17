package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// color_anthems.go — "<Color> creatures you control get +N/+M." (#5634:
// Kaysa, and Juniper Order Advocate with its "as long as this creature
// is untapped"). A layer 7c static (CR 613.4c) over every creature of
// that colour its controller controls, the source included when it is
// one. The colour is a word of the card's text (CR 612.2), so a
// text-changing effect moves it, and the creature's colour is read after
// layer 5, so a creature painted that colour counts.
//
// Append-only: add a builder, never change what one means.

// creaturesOfColorYouControlGet is the static. `color` is the colour
// letter ("G"); `while`, when set, is a condition on the source that
// must hold for the static to apply at all ("as long as this creature is
// untapped"). `label` is the printed line.
func creaturesOfColorYouControlGet(color string, power, toughness int, while func(source *game.Card) bool, label string) game.StaticAbility {
	return Labelled(b16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		if while != nil && !while(source) {
			return false
		}
		return target.IsCreature() && target.Controller == source.Controller &&
			slices.Contains(target.EffectiveColors(), source.TextColor(color))
	}, power, toughness), label)
}
