package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "a2daaf32-dbfe-4618-892e-0da24f63a44a",
		Name:            "Angel of Mercy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Angel of Mercy — you gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
