package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elven Chorus — Enchantment {3}{G}:
//
//	"You may look at the top card of your library any time.
//	 You may cast creature spells from the top of your library.
//	 Creatures you control have '{T}: Add one mana of any color.'"
//
// Three unrelated static grants, each an existing Spec slot:
// LibraryTopVisible: LibraryTopOwner is a PRIVATE look (Bolas's
// Citadel's shape), not the public LibraryTopRevealed of the Future
// Sight family — the card says "you may look", never "revealed". The
// cast permission is the Realmwalker/Oracle of Mul Daya family's
// PlayFromTopOfYourLibrary, narrowed to creature spells. The mana
// grant is the ordinary ADR 0093 shape.
//
// No simplification.
func init() {
	const grant = "elven-chorus/any-color"
	Register(Spec{
		OracleID:          "4d5df2d9-06bf-4bec-9be0-dabe75a858fb",
		Name:              "Elven Chorus",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{CreatureOnly: true},
				"Cast creature spells from the top of your library (Elven Chorus)"),
		},
		Grants: []AbilityGrant{{
			Key:  grant,
			Text: "{T}: Add one mana of any color.",
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			}},
		}},
		Static: []game.StaticAbility{
			GrantAbilities(b16CreaturesYouControl, grant),
		},
	})
}
