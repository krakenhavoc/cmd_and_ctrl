package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancient Greenwarden — Creature — Elemental {4}{G}{G}, 5/7:
//
//	"Reach
//	 You may play lands from your graveyard.
//	 If a land entering causes a triggered ability of a permanent you
//	 control to trigger, that ability triggers an additional time."
//
// Three lines of existing vocabulary. Reach rides PrintedKeywords; the
// graveyard land permission is Crucible of Worlds' standing
// CastPermission; the last line is Panharmonicon's trigger doubler
// narrowed to lands (DoublesEntering), so landfall on a permanent you
// control fires twice, and the Greenwarden's own triggers qualify as
// "a permanent you control".
//
// No simplification.
func init() {
	doubler := DoublesEntering(Land())
	doubler.Label = "Ancient Greenwarden"
	Register(Spec{
		OracleID:        "3bcf090c-e890-4a9f-a8aa-6079e4ec9947",
		Name:            "Ancient Greenwarden",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		CastPermissions: []game.CastPermission{{
			Zone:   game.ZoneGraveyard,
			Filter: game.PermissionFilter{LandsOnly: true},
			Label:  "Play a land from your graveyard (Ancient Greenwarden)",
		}},
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})
}
