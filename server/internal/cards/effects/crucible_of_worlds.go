package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crucible of Worlds — Artifact {3}:
//
//	"You may play lands from your graveyard."
//
// A standing graveyard permission, Glacierwood Siege's Sultai line
// without the chosen-word gate. A land played this way still uses the
// land drop (CR 305.2) like any other play.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "33c722cf-b4bf-431f-aefd-ee96241a7fbf",
		Name:         "Crucible of Worlds",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:   game.ZoneGraveyard,
			Filter: game.PermissionFilter{LandsOnly: true},
			Label:  "Play a land from your graveyard (Crucible of Worlds)",
		}},
	})
}
