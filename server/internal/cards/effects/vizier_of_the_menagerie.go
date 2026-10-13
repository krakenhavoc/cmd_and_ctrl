package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vizier of the Menagerie — Creature — Snake Cleric {3}{G}, 3/4:
//
//	"You may look at the top card of your library any time.
//	 You may cast creature spells from the top of your library.
//	 You can spend mana of any type to cast creature spells."
//
// The look is the private strength (LibraryTopOwner) and the
// permission is the Elven Chorus one: creature spells only, printed
// cost. The third line is Oath of Nissa's shape — an AnyColorSpend
// narrowed to casting a creature spell (CR 609.4b), from any zone, not
// only the library. "Any type" is wider than "any color": colourless is
// a type, so coloured mana also pays a {C} symbol. That is the AnyType
// kind of the grant (YouMaySpendManaAsAnyTypeToCast).
func init() {
	Register(Spec{
		OracleID:          "4f890d42-e4f1-4eed-aa6c-718865de8927",
		Name:              "Vizier of the Menagerie",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{CreatureOnly: true},
				"Cast a creature spell from the top of your library (Vizier of the Menagerie)"),
		},
		AnyColorSpend: YouMaySpendManaAsAnyTypeToCast("creature spells", "Creature"),
	})
}
