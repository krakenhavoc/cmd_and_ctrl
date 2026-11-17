package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sigarda, Font of Blessings — Legendary Creature — Angel {2}{G}{W}, 4/4:
//
//	"Flying
//	 Other permanents you control have hexproof.
//	 You may look at the top card of your library any time.
//	 You may cast Angel spells and Human spells from the top of your library."
//
// The hexproof is Privileged Position's layer-6 grant over every other
// permanent its controller controls. The look is the private strength
// (LibraryTopOwner) and the cast is a top-of-library permission whose
// filter is the union of the two creature types (CreatureTypesAny), read
// as the Mikey & Don permission reads it: spells only, so a land on top is
// not offered, and at the card's printed cost and timing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "5fe07a76-f261-4f16-8a15-571bed707261",
		Name:              "Sigarda, Font of Blessings",
		Completeness:      CompletenessFull,
		PrintedKeywords:   []string{"flying"},
		LibraryTopVisible: game.LibraryTopOwner,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller
			}, "hexproof"),
		},
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{NonLandOnly: true, CreatureTypesAny: [2]string{"Angel", "Human"}},
				"Cast an Angel or Human spell from the top of your library (Sigarda, Font of Blessings)"),
		},
	})
}
