package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wolfir Silverheart — {3}{G}{G} Creature — Wolf Warrior 4/4 (Avacyn Restored, #5570):
//
//	"Soulbond (You may pair this creature with another unpaired
//	 creature when either enters. They remain paired for as long as you
//	 control both of them.)
//	 As long as this creature is paired with another creature, each of
//	 those creatures gets +4/+4."
//
// Soulbond is the keyword's two triggers (CR 702.95a, soulbond.go).
// The last line is a layer-7c modify of both creatures of the pair
// (PairedCreaturesGet), for as long as the pair lasts: it ends when
// either creature leaves, stops being a creature or changes controller
// (CR 702.95e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dccf9497-cc7b-42df-9b46-f5ffea0c7474",
		Name:            "Wolfir Silverheart",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       Soulbond("Wolfir Silverheart"),
		Static: []game.StaticAbility{
			PairedCreaturesGet("As long as this creature is paired with another creature, each of those creatures gets +4/+4.", 4, 4),
		},
	})
}
