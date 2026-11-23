package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Druid's Familiar — {3}{G} Creature — Bear 2/2 (Avacyn Restored, #5570):
//
//	"Soulbond (You may pair this creature with another unpaired
//	 creature when either enters. They remain paired for as long as you
//	 control both of them.)
//	 As long as this creature is paired with another creature, each of
//	 those creatures gets +2/+2."
//
// Soulbond is the keyword's two triggers (CR 702.95a, soulbond.go).
// The last line is a layer-7c modify of both creatures of the pair
// (PairedCreaturesGet), for as long as the pair lasts: it ends when
// either creature leaves, stops being a creature or changes controller
// (CR 702.95e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6a5b1cd4-e597-4122-8fc7-4dce241d7c01",
		Name:            "Druid's Familiar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       effects.Soulbond("Druid's Familiar"),
		Static: []game.StaticAbility{
			effects.PairedCreaturesGet("As long as this creature is paired with another creature, each of those creatures gets +2/+2.", 2, 2),
		},
	})
}
