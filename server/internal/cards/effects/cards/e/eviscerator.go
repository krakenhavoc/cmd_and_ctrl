package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eviscerator — Creature — Phyrexian Horror {3}{B}{B}, 5/5 (Urza's
// Legacy, #5623):
//
//	"Protection from white
//	 When this creature enters, you lose 5 life."
//
// Foul Imp's enters trigger at 5 life, and protection from white as a
// printed keyword (CR 702.16a). The life loss uses the stack, so it can be
// answered, and it is loss, not damage, so protection does not stop it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "710e6d6a-caaa-4b76-82c8-e363e9438bec",
		Name:            "Eviscerator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from white"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Eviscerator — you lose 5 life", effects.Do(effects.GainLife{Amount: -5})),
		},
	})
}
