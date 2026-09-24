package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gemhide Sliver — Creature — Sliver {1}{G}, 1/1:
//
//	"All Slivers have "{T}: Add one mana of any color.""
//
// ALL Slivers — every player's, and this one. The grant is each
// Sliver's own ability, so an opponent's Sliver taps for its
// controller's mana, never for yours (CR 602.2).
//
// No simplification.
const gemhideSliverGrant = "gemhide-sliver/any-color"

func init() {
	Register(Spec{
		OracleID:     "2c09ca09-8e62-4fe3-9b3d-61573dd2ffbc",
		Name:         "Gemhide Sliver",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{AnyColorManaGrant(gemhideSliverGrant)},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}}, gemhideSliverGrant),
		},
	})
}
