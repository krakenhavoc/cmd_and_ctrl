package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battle Mastery — Enchantment — Aura for {2}{W}:
//
//	"Enchant creature
//	 Enchanted creature has double strike."
//
// The Aura counterpart of Basilisk Collar: an enchant clause, one
// layer 6 grant, and nothing else. It is here because double strike
// is the keyword that makes the rest of the attachment surface pay
// off twice — the Warhammer's +3/+0, a Sword's combat-damage trigger,
// Skullclamp's carrier — and because it is the second Aura in the
// catalog that is not a control-changer, which is what the CR 704.5n
// branch needed to be exercised by something other than Rancor.
//
// Double strike is honoured: the combat damage step runs twice for
// it, and a creature with both first strike and double strike still
// gets exactly two hits.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a06c7c1a-8534-40fb-8bb8-59f8f2530567",
		Name:         "Battle Mastery",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			GrantToAttached("double strike"),
		},
	})
}
