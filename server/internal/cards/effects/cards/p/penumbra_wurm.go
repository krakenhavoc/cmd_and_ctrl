package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Penumbra Wurm — Creature — Wurm {5}{G}{G}, 6/6 (Apocalypse, #5614):
//
//	"Trample
//	 When this creature dies, create a 6/6 black Wurm creature token with
//	 trample."
//
// Penumbra Kavu's dies trigger with the printed 6/6 black Wurm token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "abccb19f-098b-4453-b7bc-838ad5914ebe",
		Name:            "Penumbra Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisDies("Penumbra Wurm — create a 6/6 black Wurm creature token with trample",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("6/6 black Wurm with trample"), N: 1})),
				game.Purpose{Tokens: 1}),
		},
	})
}
