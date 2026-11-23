package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Mauler — {1}{R} Creature — Human Berserker 2/1 (Avacyn Restored, #5570):
//
//	"Soulbond (You may pair this creature with another unpaired
//	 creature when either enters. They remain paired for as long as you
//	 control both of them.)
//	 As long as this creature is paired with another creature, both
//	 creatures have haste."
//
// Soulbond is the keyword's two triggers (CR 702.95a, soulbond.go).
// The last line is a layer-6 keyword grant to both creatures of the
// pair (PairedCreaturesHaveKeyword), for as long as the pair lasts: it
// ends when either creature leaves, stops being a creature or changes
// controller (CR 702.95e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6c5664b1-fbb5-477e-80fd-cf2285014a10",
		Name:            "Lightning Mauler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       effects.Soulbond("Lightning Mauler"),
		Static: []game.StaticAbility{
			effects.PairedCreaturesHaveKeyword("As long as this creature is paired with another creature, both creatures have haste.", "haste"),
		},
	})
}
