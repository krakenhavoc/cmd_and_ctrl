package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ee4b7aca-31dd-4369-8e85-b300892b9b41",
		Name:         "Perennial Behemoth",
		Completeness: effects.CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:   game.ZoneGraveyard,
			Filter: game.PermissionFilter{LandsOnly: true},
			Label:  "Play a land from your graveyard (Perennial Behemoth)",
		}},
		Activated: []effects.ActivatedAbility{effects.Unearth("{G}{G}")},
	})
}
