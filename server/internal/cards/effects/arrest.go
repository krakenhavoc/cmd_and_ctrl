package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arrest — Enchantment — Aura for {2}{W}:
//
//	"Enchant creature
//	 Enchanted creature can't attack or block, and its activated
//	 abilities can't be activated."
//
// Pacifism plus the clause that makes it the one you actually want
// in Commander: the format's threats are rarely the ones that attack.
// A pacified Krenko still makes Goblins; an arrested one does not.
//
// The activation half is CR 602.5 and it is absolute here —
// "its activated abilities" with no carve-out, so a Birds of
// Paradise under Arrest taps for nothing. That is the difference
// from Faith's Fetters, which spells out an exception for mana
// abilities, and it is why the engine carries two bits rather than
// one (game.CantActivate / game.CantActivateMana). Both are checked
// at the activation entry points before any cost is validated, so a
// refused activation costs the player nothing.
//
// Loyalty abilities are activated abilities (CR 606.1) and are
// covered by the same bit — irrelevant on Arrest, which can only
// enchant a creature, and the reason Faith's Fetters is worth
// having next to it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "81728b98-8cf9-4734-a318-69184bb4d15c",
		Name:         "Arrest",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			RestrictAttached(
				game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana,
			),
		},
	})
}
