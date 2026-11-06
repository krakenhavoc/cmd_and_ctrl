package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chitinous Graspling — Creature — Shapeshifter {3}{G/U}, 3/4 (Lorwyn
// Eclipsed, #4458):
//
//	"Changeling (This card is every creature type.)
//	 Reach"
//
// Two printed keywords and nothing else, as Barkform Harvester declares
// them: changeling is the every-creature-type marker HasSubtype reads in
// every zone, and reach is read by the block-pair check.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "930035c6-cde7-40ac-98ef-0840812037db",
		Name:            "Chitinous Graspling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling, "reach"},
	})
}
