package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wingcrafter — {U} Creature — Human Wizard 1/1 (Avacyn Restored, #5570):
//
//	"Soulbond (You may pair this creature with another unpaired
//	 creature when either enters. They remain paired for as long as you
//	 control both of them.)
//	 As long as this creature is paired with another creature, both
//	 creatures have flying."
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
		OracleID:        "55d1f44f-8554-49e7-b36f-92682926e6d0",
		Name:            "Wingcrafter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       effects.Soulbond("Wingcrafter"),
		Static: []game.StaticAbility{
			effects.PairedCreaturesHaveKeyword("As long as this creature is paired with another creature, both creatures have flying.", "flying"),
		},
	})
}
