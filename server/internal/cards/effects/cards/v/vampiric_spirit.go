package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampiric Spirit — Creature — Spirit {2}{B}{B}, 4/3 (Eighth Edition, S82 OLD
// batch 609, #5607):
//
//	"Flying
//	 When this creature enters, you lose 4 life."
//
// The life loss is not damage (CR 119.3) and is not optional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dbd1fc9c-6316-4c89-a971-f2e08ca291a4",
		Name:            "Vampiric Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Vampiric Spirit — you lose 4 life", effects.Do(effects.YouLoseLife(4))),
		},
	})
}
