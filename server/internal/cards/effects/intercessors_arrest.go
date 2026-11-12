package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Intercessor's Arrest — Enchantment — Aura {2}{W} (Kamigawa: Neon
// Dynasty, #5673):
//
//	"Enchant permanent
//	 Enchanted permanent can't attack, block, or crew Vehicles. Its
//	 activated abilities can't be activated unless they're mana
//	 abilities."
//
// Faith's Fetters' lock with one more word. "Can't crew Vehicles" is
// the CantCrew restriction (CR 702.122d): the enchanted creature can't
// be tapped to pay a Vehicle's crew cost, so the crew validator refuses
// it and neither the crew picker nor a bot is offered it. A saddle cost
// and every other "tap a creature you control" cost still take it; the
// card stops crewing and nothing else.
//
// "Enchant permanent" reaches a Vehicle too. There the crew half does
// nothing until the Vehicle is a creature, and the activation half is
// what matters: crew is an activated ability and not a mana ability, so
// the enchanted Vehicle can't be crewed at all (CR 602.5).
//
// Mana abilities are spared (CantActivate without CantActivateMana), as
// printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d75357f4-d439-48ff-8fd7-52a02f07cc81",
		Name:         "Intercessor's Arrest",
		Completeness: CompletenessFull,
		Targets:      EnchantPermanent(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			RestrictAttached(game.CantAttackOrBlock | game.CantCrew | game.CantActivate),
		},
	})
}
