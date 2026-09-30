package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const imprisonedInTheMoonGrant = "imprisoned-in-the-moon/tap-for-colorless"

// Imprisoned in the Moon — Enchantment — Aura {2}{U}:
//
//	"Enchant creature, land, or planeswalker
//	 Enchanted permanent is a colorless land with '{T}: Add {C}' and
//	 loses all other card types and abilities."
//
// Three layers, declared in printed order, the way Darksteel Mutation
// is:
//
//   - layer 4: the permanent is a land and nothing else. A REPLACEMENT
//     of the card types, and of the subtypes with them — this is not
//     "a Forest", so a basic land loses its land type and the
//     intrinsic mana ability that came with it. Supertypes survive
//     (CR 205.4).
//   - layer 5: colorless.
//   - layer 6: loses all abilities, and has "{T}: Add {C}". One
//     effect, so the removal empties the slice before the grant
//     appends to it (ADR 0093 Decision 3) — the granted mana ability
//     is the only ability the permanent has.
//
// The enchant clause is the wide one because the Aura may go on any of
// the three; the CR 704.5m re-check runs the same predicate, and a
// creature that becomes a land is still a legal host.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "339a7418-9a29-4eae-a3c2-ea590d175936",
		Name:         "Imprisoned in the Moon",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("enchant creature, land, or planeswalker",
			Or(Creature(), Land(), Planeswalker())),
		Grants: []AbilityGrant{
			TapForManaGrant(imprisonedInTheMoonGrant, "{C}", "Add {C}", "{T}: Add {C}."),
		},
		Static: []game.StaticAbility{
			SetAttachedTypes([]string{"Land"}, nil),
			SetAttachedColors(),
			imprisonedInTheMoonAbilities(),
		},
	})
}

// imprisonedInTheMoonAbilities is "loses all other abilities" and "has
// '{T}: Add {C}'" as the ONE layer-6 effect the card prints.
func imprisonedInTheMoonAbilities() game.StaticAbility {
	s := LoseAllAbilities()
	s.GrantAbilities = []string{imprisonedInTheMoonGrant}
	return s
}
