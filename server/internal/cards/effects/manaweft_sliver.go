package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Manaweft Sliver — Creature — Sliver {1}{G}, 1/1:
//
//	"Sliver creatures you control have "{T}: Add one mana of any
//	 color.""
//
// Gemhide Sliver's grant with a controller clause: only your Slivers.
//
// No simplification.
const manaweftSliverGrant = "manaweft-sliver/any-color"

func init() {
	Register(Spec{
		OracleID:     "bd47398d-da35-4a09-8754-771af91b14f4",
		Name:         "Manaweft Sliver",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{AnyColorManaGrant(manaweftSliverGrant)},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, manaweftSliverGrant),
		},
	})
}
