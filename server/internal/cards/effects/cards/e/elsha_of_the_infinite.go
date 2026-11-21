package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elsha of the Infinite — Legendary Creature — Djinn Monk {2}{U}{R}{W},
// 3/3 (Commander 2019, #5517):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 You may look at the top card of your library any time.
//	 You may cast noncreature spells from the top of your library. If you
//	 cast a spell this way, you may cast it as though it had flash."
//
// Prowess is derived by the engine from the keyword. The look is private
// (LibraryTopOwner, Augur of Autumn's). The cast is a standing permission
// over the top card of your library (PlayFromTopOfYourLibrary), narrowed to
// noncreature spells and to CASTING, so a land on top cannot be played
// through it (a land is not a spell, CR 305.1). The permission carries
// TimingFlash, so a spell cast through it may be cast any time you could
// cast an instant, and a spell cast from the top some other way does not.
//
// No simplification.
func init() {
	perm := effects.PlayFromTopOfYourLibrary(game.PermissionFilter{NoncreatureOnly: true},
		"Cast a noncreature spell from the top of your library, as though it had flash (Elsha of the Infinite)")
	perm.CastOnly = true
	perm.Timing = game.TimingFlash
	effects.Register(effects.Spec{
		OracleID:          "da339200-ed26-42e4-b24f-b96bc4daef7c",
		Name:              "Elsha of the Infinite",
		Completeness:      effects.CompletenessFull,
		PrintedKeywords:   []string{"prowess"},
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions:   []game.CastPermission{perm},
	})
}
