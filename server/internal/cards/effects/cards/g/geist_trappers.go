package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geist Trappers — {4}{G} Creature — Human Warrior 3/5 (Avacyn Restored, #5570):
//
//	"Soulbond (You may pair this creature with another unpaired
//	 creature when either enters. They remain paired for as long as you
//	 control both of them.)
//	 As long as this creature is paired with another creature, both
//	 creatures have reach."
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
		OracleID:        "7116da4a-96c1-4183-b123-29c6b70ef543",
		Name:            "Geist Trappers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       effects.Soulbond("Geist Trappers"),
		Static: []game.StaticAbility{
			effects.PairedCreaturesHaveKeyword("As long as this creature is paired with another creature, both creatures have reach.", "reach"),
		},
	})
}
