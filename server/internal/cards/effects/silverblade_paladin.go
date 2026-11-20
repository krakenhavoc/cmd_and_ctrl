package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silverblade Paladin — {1}{W}{W} Creature — Human Knight 2/2 (Avacyn Restored, #5570):
//
//	"Soulbond (You may pair this creature with another unpaired
//	 creature when either enters. They remain paired for as long as you
//	 control both of them.)
//	 As long as this creature is paired with another creature, both
//	 creatures have double strike."
//
// Soulbond is the keyword's two triggers (CR 702.95a, soulbond.go).
// The last line is a layer-6 keyword grant to both creatures of the
// pair (PairedCreaturesHaveKeyword), for as long as the pair lasts: it
// ends when either creature leaves, stops being a creature or changes
// controller (CR 702.95e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ded0316f-5eed-4a80-91ad-ca236fd16f01",
		Name:            "Silverblade Paladin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       Soulbond("Silverblade Paladin"),
		Static: []game.StaticAbility{
			PairedCreaturesHaveKeyword("As long as this creature is paired with another creature, both creatures have double strike.", "double strike"),
		},
	})
}
