package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thundermane Dragon — Creature — Dragon {3}{R}, 4/4 (Tarkir: Dragonstorm
// Commander):
//
//	"Flying
//	 You may look at the top card of your library any time.
//	 You may cast creature spells with power 4 or greater from the top of your
//	 library. If you cast a creature spell this way, it gains haste until end
//	 of turn."
//
// The look is the shared top-of-library clause (LibraryTopOwner). The cast is
// a standing library permission (CR 401.5, derived off the battlefield on every
// query) narrowed to creature spells whose power, judged on the face being cast
// with a `*` worked out by the game, is 4 or greater
// (PermissionFilter.PowerAtLeast). It adds no price and no timing, so the
// spell is cast for its own costs at its own timing (the ruling: a main phase
// with an empty stack, or any time for flash). The haste rides the permission
// (CastPermission.GainsKeywords), is read as the cast is made and lasts to the
// end of that turn (GainsUntilEndOfTurn). It is the Dragon's controller's
// permission only, and goes with the Dragon.
//
// Seams: PermissionFilter.PowerAtLeast, CastPermission.GainsUntilEndOfTurn.
func init() {
	Register(Spec{
		OracleID:          "a1602bd0-f346-43cf-86ec-1dd300a7f2ed",
		Name:              "Thundermane Dragon",
		Completeness:      CompletenessFull,
		PrintedKeywords:   []string{"flying"},
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{{
			Zone:                game.ZoneLibrary,
			TopOfLibraryOnly:    true,
			CastOnly:            true,
			Filter:              game.PermissionFilter{CreatureOnly: true, PowerAtLeastCapped: true, PowerAtLeast: 4},
			GainsKeywords:       []string{"haste"},
			GainsUntilEndOfTurn: true,
			Label:               "Cast creature spells with power 4 or greater from the top of your library; they gain haste until end of turn (Thundermane Dragon)",
		}},
	})
}
