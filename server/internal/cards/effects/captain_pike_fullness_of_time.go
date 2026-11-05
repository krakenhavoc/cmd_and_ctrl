package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain Pike, Fullness of Time — Legendary Creature — Human Officer
// Scout {1}{G}{G}, */* (Star Trek, #4973):
//
//	"Strange new worlds — Captain Pike's power and toughness are each
//	 equal to the number of differently named lands you control.
//	 You may look at the top card of your library any time.
//	 You may play lands from the top of your library."
//
// "Strange new worlds" is an ability word (CR 207.2c); the first line is
// a characteristic-defining ability (CR 604.3a) counted in layer 7a, each
// land name once. The look is Bolas's Citadel's private visibility
// (LibraryTopOwner: the top card is shown to its owner, not revealed);
// the play is Courser of Kruphix's lands-only permission, which still
// spends a land drop.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "745633e6-8635-4d99-8a4e-edc5d500bb3d",
		Name:              "Captain Pike, Fullness of Time",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		Static: []game.StaticAbility{
			trekB19DifferentlyNamedLandsPT("Strange new worlds — Captain Pike's power and toughness are each equal to the number of differently named lands you control."),
		},
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{LandsOnly: true},
				"Play a land from the top of your library (Captain Pike, Fullness of Time)"),
		},
	})
}
