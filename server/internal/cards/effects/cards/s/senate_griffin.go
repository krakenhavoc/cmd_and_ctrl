package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Senate Griffin — Creature — Griffin {2}{W/U}{W/U}, 3/2 (Ravnica
// Allegiance, #5521):
//
//	"Flying
//	 When this creature enters, scry 1."
//
// Silver Raven's two lines: flying is printed data (CR 702.9b), and the
// enters trigger scries 1 (CR 701.22a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7ddf9c12-8f75-481c-a218-d42f925c6503",
		Name:            "Senate Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Senate Griffin — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
