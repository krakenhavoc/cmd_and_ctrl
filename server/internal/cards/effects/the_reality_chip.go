package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Reality Chip — Legendary Artifact Creature — Equipment Jellyfish
// {1}{U}, 0/4:
//
//	"You may look at the top card of your library any time.
//	 As long as The Reality Chip is attached to a creature, you may play
//	 lands and cast spells from the top of your library.
//	 Reconfigure {2}{U}"
//
// The look is the private strength of ADR 0066's library-top rule
// (LibraryTopOwner, Bolas's Citadel's). The play permission is Magus of
// the Future's, gated on the Chip being attached to a creature
// (GatedCastPermissions, the Fortune Teller's Talent shape). Reconfigure
// is reconfigure.go (#2639).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "751f22bc-bcea-4213-a37b-b5f72448a4c4",
		Name:              "The Reality Chip",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		GatedCastPermissions: []game.CastPermissionGate{{
			Permission: PlayFromTopOfYourLibrary(game.PermissionFilter{},
				"Play a land or cast a spell from the top of your library (The Reality Chip)"),
			Condition: func(g *game.Game, _, source uuid.UUID) bool {
				return g.AttachedToACreatureForEffect(source)
			},
		}},
		Activated: Reconfigure("{2}{U}"),
	})
}
