package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cursed Totem — {2} Artifact:
//
//	"Activated abilities of creatures can't be activated."
//
// #1210's headline card and the simplest possible activation
// restriction (CR 602.5): one predicate over the object whose
// ability is being announced. Nothing is stored on any permanent —
// the restriction is DERIVED from this artifact being on the
// battlefield, so it lifts the moment Cursed Totem leaves and a
// second copy changes nothing. Exactly Rule of Law's shape, one gate
// over.
//
// NO MANA EXEMPTION, and that is the entire card. Collector Ouphe
// below stops artifacts and Linvala stops your opponents' creatures;
// what makes Cursed Totem the card that beats an elf-ball deck is
// that "activated abilities of creatures" reaches a Birds of
// Paradise's {T}: Add and a Llanowar Elves' too. CR 605.1a makes a
// mana ability an activated ability, and this clause prints no
// "unless", so `exemptMana` is false and the ONE gate refuses both
// paths — including the auto-tapper, which does not go through
// ActivateManaAbility at all and would otherwise have minted the mana
// anyway.
//
// "Creatures" is every creature on the battlefield, its controller's
// included. The card is symmetrical and the asymmetrical one says so
// (Linvala's "your opponents control"), which is why they are two
// constructors and not a flag.
//
// The predicate reads the LAYERED type line (Card.IsCreature is layer
// 4's output since ADR 0039), so a Darksteel Mutation'd Sol Ring is a
// creature for this purpose and its equip really is shut off — which
// is the printed rule, not a convenience.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6225a704-430a-4f56-ad87-0e8d87f285f5",
		Name:         "Cursed Totem",
		Completeness: CompletenessFull,
		ActivationRestrictions: []game.ActivationRestriction{
			SourcesCantActivate(
				"Cursed Totem — activated abilities of creatures can't be activated.",
				Creature(), false),
		},
	})
}
