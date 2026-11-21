package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harried Spearguard — Creature — Human Soldier {R}, 1/1 (Wilds of
// Eldraine, #4471):
//
//	"Haste
//	 When this creature dies, create a 1/1 black Rat creature token with
//	 "This token can't block.""
//
// Haste is the engine's keyword (CR 702.10). The dies trigger makes
// Twisted Sewer-Witch's Rat, the catalog token whose own text is "This
// token can't block", under the trigger's controller (Edgewall Pack's
// token, on a death rather than an entry).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e3542d0f-fcee-4df7-aa95-b55138caa2ab",
		Name:            "Harried Spearguard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Harried Spearguard — create a 1/1 black Rat creature token with \"This token can't block.\"",
				effects.Do(effects.CreateToken{Template: effects.CantBlockRatToken(), N: 1})),
		},
	})
}
