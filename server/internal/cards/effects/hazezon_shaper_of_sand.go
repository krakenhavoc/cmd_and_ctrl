package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hazezon, Shaper of Sand:
//
//	"Desertwalk
//	 You may play Desert lands from your graveyard.
//	 Whenever a Desert you control enters, create two 1/1 red, green, and
//	 white Sand Warrior creature tokens."
//
// Desertwalk is CR 702.14a's landwalk with the Desert land type. The
// permission is Crucible of Worlds' narrowed to Deserts; a land played
// this way uses the land drop (CR 305.2). Hazezon's own trigger is for
// each Desert, so two Deserts entering together make four tokens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "66b16cb3-c2b1-447c-a4b0-a67326c99f20",
		Name:            "Hazezon, Shaper of Sand",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"desertwalk"},
		CastPermissions: []game.CastPermission{{
			Zone:   game.ZoneGraveyard,
			Filter: game.PermissionFilter{LandsOnly: true, Subtype: "Desert"},
			Label:  "Play a Desert from your graveyard (Hazezon, Shaper of Sand)",
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, cmdr2S14SubtypeLandEntersUnderYourControl("Desert"),
				"Hazezon, Shaper of Sand — create two 1/1 red, green, and white Sand Warrior creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 red, green, and white Sand Warrior"), N: 2})),
		},
	})
}
