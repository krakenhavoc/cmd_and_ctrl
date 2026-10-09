package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mystic Forge — Artifact {4}:
//
//	"You may look at the top card of your library any time.
//	 You may cast artifact spells and colorless spells from the top of
//	 your library.
//	 {T}, Pay 1 life: Exile the top card of your library."
//
// The look is private (LibraryTopOwner, CR 401.5) and the permission is
// the Realmwalker shape with the filter's artifact-or-colorless OR
// (ArtifactOrColorlessOnly) and NonLandOnly: a land on top is visible
// and unplayable, and a colorless land is not a spell.
//
// The exile ability pays {T} and the life at announcement (CR 602.2b)
// and exiles whatever is on top when it RESOLVES, so a card cast off
// the top in response means the next one goes. The exile is public,
// face up, and is not a draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "994bc16d-fdc6-475c-96df-beeb4d5aa03e",
		Name:              "Mystic Forge",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{NonLandOnly: true, ArtifactOrColorlessOnly: true},
				"Cast an artifact or colorless spell from the top of your library (Mystic Forge)"),
		},
		Activated: []ActivatedAbility{{
			Label: "{T}, Pay 1 life: Exile the top card of your library.",
			Cost:  Plus(TapCost(), PayLife(1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return MillToZone{Player: ctx.Controller(), N: 1, To: game.ZoneExile}.Apply(ctx)
			},
		}},
	})
}
