package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hanweir Lancer — {2}{R} Creature — Human Knight 2/2 (Avacyn Restored, #5570):
//
//	"Soulbond (You may pair this creature with another unpaired
//	 creature when either enters. They remain paired for as long as you
//	 control both of them.)
//	 As long as this creature is paired with another creature, both
//	 creatures have first strike."
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
		OracleID:        "5fa9de04-6b70-4162-9ffb-6334277a4f30",
		Name:            "Hanweir Lancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       Soulbond("Hanweir Lancer"),
		Static: []game.StaticAbility{
			PairedCreaturesHaveKeyword("As long as this creature is paired with another creature, both creatures have first strike.", "first strike"),
		},
	})
}
