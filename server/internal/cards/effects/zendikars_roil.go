package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zendikar's Roil — Enchantment {3}{G}{G} (EDHREC rank 2119):
//
//	"Landfall — Whenever a land you control enters, create a 2/2
//	 green Elemental creature token."
//
// Rampaging Baloths on an enchantment, half the size. Landfall is
// the ETB-filtered trigger every landfall card in the catalog uses
// (Avenger of Zendikar, Tireless Provisioner): a land entering under
// the controller's control — played, fetched, or reanimated — makes
// one Elemental.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a842cc2b-52eb-4dc5-86c6-6575c2ed913d",
		Name:         "Zendikar's Roil",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Zendikar's Roil — create a 2/2 Elemental (landfall)", Do(CreateToken{Template: b19GreenElementalToken(), N: 1})),
		},
	})
}
