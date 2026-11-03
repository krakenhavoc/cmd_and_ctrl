package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sky Tether — Enchantment — Aura {W} (Jumpstart 2020):
//
//	"Enchant creature
//	 Enchanted creature has defender and loses flying."
//
// Two layer-6 statics on the enchanted creature, in printed order: the
// defender grant (read by attack legality) and Colossus Hammer's
// RemoveFromAttached for "loses flying" (read by block legality). As
// on the Hammer, a flying grant with a later timestamp lands after the
// removal and survives it (CR 613.7); that is the layer system, not a
// leak.
//
// No simplification. Deck request #4894 (Jumpstart Azorius).
func init() {
	Register(Spec{
		OracleID:     "06548af1-4fda-435c-ac32-bbf7afafb34e",
		Name:         "Sky Tether",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			GrantToAttached("defender"),
			RemoveFromAttached("flying"),
		},
	})
}
