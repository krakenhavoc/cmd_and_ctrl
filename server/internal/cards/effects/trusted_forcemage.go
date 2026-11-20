package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trusted Forcemage — {2}{G} Creature — Human Shaman 2/2 (Avacyn Restored, #5570):
//
//	"Soulbond (You may pair this creature with another unpaired
//	 creature when either enters. They remain paired for as long as you
//	 control both of them.)
//	 As long as this creature is paired with another creature, each of
//	 those creatures gets +1/+1."
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
		OracleID:        "663ef3c9-f897-4875-b029-82ee6672db00",
		Name:            "Trusted Forcemage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       Soulbond("Trusted Forcemage"),
		Static: []game.StaticAbility{
			PairedCreaturesGet("As long as this creature is paired with another creature, each of those creatures gets +1/+1.", 1, 1),
		},
	})
}
