package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Perennial Behemoth:
//
//	"You may play lands from your graveyard.
//	 Unearth {G}{G}"
//
// Crucible of Worlds on a body (a land played this way still uses the
// land drop), and the unearth keyword's whole ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ee4b7aca-31dd-4369-8e85-b300892b9b41",
		Name:         "Perennial Behemoth",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:   game.ZoneGraveyard,
			Filter: game.PermissionFilter{LandsOnly: true},
			Label:  "Play a land from your graveyard (Perennial Behemoth)",
		}},
		Activated: []ActivatedAbility{Unearth("{G}{G}")},
	})
}
