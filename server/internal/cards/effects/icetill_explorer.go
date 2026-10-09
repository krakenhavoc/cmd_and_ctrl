package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Icetill Explorer — Creature — Insect Scout {2}{G}{G}, 2/4:
//
//	"You may play an additional land on each of your turns.
//	 You may play lands from your graveyard.
//	 Landfall — Whenever a land you control enters, mill a card."
//
// Three printed lines, three existing seams: Exploration's extra land
// drop (Spec.AdditionalLandPlays), Crucible of Worlds' graveyard play
// permission, and the shared landfall trigger over MillCards. A land
// played from the graveyard still uses a land drop (CR 305.2), and
// the mill can put the next land into the graveyard for the
// permission to find.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "109cdefd-e8cc-4ac7-b6ba-2cfdef8d780f",
		Name:                "Icetill Explorer",
		Completeness:        CompletenessFull,
		AdditionalLandPlays: 1,
		Purpose:             game.Purpose{ExtraLandDrops: 1}, // #2678: the bot reads the extra drop
		CastPermissions: []game.CastPermission{{
			Zone:   game.ZoneGraveyard,
			Filter: game.PermissionFilter{LandsOnly: true},
			Label:  "Play a land from your graveyard (Icetill Explorer)",
		}},
		Triggered: []game.TriggeredAbility{
			Landfall("Icetill Explorer — mill a card", Do(MillCards{N: 1})),
		},
	})
}
