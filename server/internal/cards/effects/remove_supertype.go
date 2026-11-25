package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// remove_supertype.go — #6199: "is no longer <supertype>" (CR 613.1d,
// layer 4). A printed static is RemovesSupertype; a resolved effect on a
// target is a ScopedEffectFor with game.RemoveSupertypesMod:
//
//	Static: []game.StaticAbility{AllLandsAreNoLongerSupertype("Snow")}  // Melting: "All lands are no longer snow."
//	ScopedEffectFor{Target: id, Mods: []game.Mod{game.RemoveSupertypesMod("Snow")},
//	    Duration: game.IndefiniteDuration(), Label: …}                 // Arcum's Weathervane
//
// "Target nonsnow basic land becomes snow" is the same record with
// game.AddSupertypesMod("Snow"). Both apply in layer 4 in timestamp order
// (CR 613.7).

// AllLandsAreNoLongerSupertype is "All lands are no longer <supertype>"
// as a printed static: a layer-4 removal of the supertype from every
// land on the battlefield, whoever controls it.
func AllLandsAreNoLongerSupertype(supertype string) game.StaticAbility {
	return game.StaticAbility{
		Label: "All lands are no longer " + strings.ToLower(supertype) + ".",
		Layer: game.Layer4Type,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsLand()
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.RemoveSupertypes([]string{supertype})
		},
	}
}
