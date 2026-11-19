package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lightning Mauler — Creature — Human Berserker {1}{R}, 2/1 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"Soulbond (You may pair this creature with another unpaired creature
//	 when either enters. They remain paired for as long as you control
//	 both of them.)
//	 As long as this creature is paired with another creature, both
//	 creatures have haste."
//
// Soulbond is the keyword's two triggers (CR 702.95a, soulbond.go). The
// second line is a layer-6 keyword grant to the Mauler and its partner,
// and it ends with the pair (CR 702.95e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6c5664b1-fbb5-477e-80fd-cf2285014a10",
		Name:            "Lightning Mauler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       Soulbond("Lightning Mauler"),
		Static: []game.StaticAbility{
			pairedCreaturesHaveKeyword("As long as this creature is paired with another creature, both creatures have haste.", "haste"),
		},
	})
}
