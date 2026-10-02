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
// only the library. "Any type" is slightly wider than "any color": it
// also lets coloured mana pay a {C} symbol. The engine's spend grant
// does not reach that half (CR 106.1b; only a creature that costs {C}
// shows it), which is the caveat.
func init() {
	Register(Spec{
		OracleID:          "4f890d42-e4f1-4eed-aa6c-718865de8927",
		Name:              "Vizier of the Menagerie",
		Completeness:      CompletenessCaveats,
		Caveats:           []string{"Mana of any type is only spent as any colour on creature spells — coloured mana still can't pay a {C} cost symbol."},
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{CreatureOnly: true},
				"Cast a creature spell from the top of your library (Vizier of the Menagerie)"),
		},
		AnyColorSpend: YouMaySpendManaAsAnyColorToCast("creature spells", "Creature"),
	})
}
