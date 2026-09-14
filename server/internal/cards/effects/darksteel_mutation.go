package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Darksteel Mutation — Enchantment — Aura for {1}{W}:
//
//	"Enchant creature
//	 Enchanted creature is an Insect artifact creature with base
//	 power and toughness 0/1 and has indestructible, and it loses
//	 all other abilities, card types, and creature types."
//
// The premier white answer to a commander in a format where killing
// one does nothing: the creature never goes to the command zone, so
// it never comes back, and it is indestructible so its controller
// cannot even sacrifice it to a wrath. It is also the card that made
// "loses all abilities" worth building, because half the commanders
// worth pointing it at do their work from an ability rather than
// from combat.
//
// Four clauses, four CR 613 layers, declared in printed order:
//
//   - layer 4 — "is an Insect artifact creature ... loses all other
//     card types and creature types". A REPLACEMENT of the type
//     line, not an addition. Supertypes survive: a legendary
//     commander under this is still legendary, so a second copy
//     still dies to the legend rule.
//   - layer 6 — "has indestructible, and it loses all other
//     abilities". One effect that removes and grants at once, which
//     is why the keep rides on LoseAllAbilities rather than sitting
//     in a separate GrantToAttached: a separate grant would be a
//     separate effect with the same timestamp, and the ordering
//     between two same-timestamp effects is gather order, which is
//     not a thing a card file should have to reason about.
//   - layer 7b — "base power and toughness 0/1". A SET, so anthems
//     (7c) and +1/+1 counters (7d) still apply on top. Printed
//     behaviour, and the reason Mutating your own creature with
//     counters on it is a decision rather than a no-op.
//
// What survives, and it is worth stating because the engine now
// enforces it: the creature keeps its name, its counters, its
// controller, its owner, its attachments both ways, and its place on
// the battlefield. CR 613.1f removes abilities, not objects.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "05a4f8ff-49da-42af-add5-6248c4b0644b",
		Name:         "Darksteel Mutation",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			SetAttachedTypes([]string{"Artifact", "Creature"}, []string{"Insect"}),
			LoseAllAbilities("indestructible"),
			SetAttachedBasePT(0, 1),
		},
	})
}
