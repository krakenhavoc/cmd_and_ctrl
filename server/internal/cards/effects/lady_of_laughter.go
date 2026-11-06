package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lady of Laughter — Creature — Faerie Noble {3}{W}{W}, 4/5 (Wilds of
// Eldraine, #4471):
//
//	"Flying
//	 Celebration — At the beginning of your end step, if two or more
//	 nonland permanents entered the battlefield under your control this
//	 turn, draw a card."
//
// The end-step trigger reads the celebration tally (celebration.go) as an
// intervening "if" (CR 603.4, AtYourEndStepIfCelebration): it triggers only
// at its controller's own end step and only when two or more nonland
// permanents entered under their control that turn, the Lady included if
// she entered that turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "356a85d9-2570-448c-a455-f4ce915a604f",
		Name:            "Lady of Laughter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfCelebration("Lady of Laughter — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
