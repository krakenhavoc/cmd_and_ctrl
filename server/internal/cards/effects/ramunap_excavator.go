package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ramunap Excavator — Creature — Snake Cleric {2}{G}, 2/3 (EDHREC top
// 1,000, #3654):
//
//	"You may play lands from your graveyard."
//
// Crucible of Worlds on a body: a standing graveyard permission for
// lands. A land played this way still uses the land drop.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f819ba4-52ef-4fdd-8e4c-5ae3b2f44db5",
		Name:         "Ramunap Excavator",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:   game.ZoneGraveyard,
			Filter: game.PermissionFilter{LandsOnly: true},
			Label:  "Play a land from your graveyard (Ramunap Excavator)",
		}},
	})
}
