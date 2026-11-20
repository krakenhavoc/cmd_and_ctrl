package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elgaud Shieldmate — {3}{U} Creature — Human Soldier 2/3 (Avacyn Restored, #5570):
//
//	"Soulbond (You may pair this creature with another unpaired
//	 creature when either enters. They remain paired for as long as you
//	 control both of them.)
//	 As long as this creature is paired with another creature, both
//	 creatures have hexproof. (They can't be the targets of spells or
//	 abilities your opponents control.)"
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
		OracleID:        "a85f7e34-7783-474c-ac96-9b075b296b8f",
		Name:            "Elgaud Shieldmate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       Soulbond("Elgaud Shieldmate"),
		Static: []game.StaticAbility{
			PairedCreaturesHaveKeyword("As long as this creature is paired with another creature, both creatures have hexproof.", "hexproof"),
		},
	})
}
