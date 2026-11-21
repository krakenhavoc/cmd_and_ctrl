package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angel of Mercy — Creature — Angel {4}{W}, 3/3 (Jumpstart Selesnya
// deck request, #4893):
//
//	"Flying
//	 When this creature enters, you gain 3 life."
//
// The enters trigger uses the stack, so it can be answered; the life goes
// to the trigger's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a2daaf32-dbfe-4618-892e-0da24f63a44a",
		Name:            "Angel of Mercy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Angel of Mercy — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
