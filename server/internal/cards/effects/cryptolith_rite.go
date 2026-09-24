package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cryptolith Rite — Enchantment {1}{G}:
//
//	"Creatures you control have "{T}: Add one mana of any color.""
//
// The ADR 0093 headline card: a layer-6 grant of a catalog bundle to
// every creature you control (CR 113.10). The ability is each
// creature's own — its controller activates it, {T} taps the
// creature, and CR 302.6 keeps a creature that entered this turn from
// using it unless it has haste. The auto-tapper plans a creature's
// granted mana only as a last resort (ADR 0093 Decision 6), so an
// auto-paid cast does not tap the creatures you meant to attack with.
//
// No simplification.
const cryptolithRiteGrant = "cryptolith-rite/any-color"

func init() {
	Register(Spec{
		OracleID:     "043f869d-b11c-4c0d-9591-2bf0df7bde55",
		Name:         "Cryptolith Rite",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{AnyColorManaGrant(cryptolithRiteGrant)},
		Static:       []game.StaticAbility{GrantAbilities(b16CreaturesYouControl, cryptolithRiteGrant)},
	})
}
