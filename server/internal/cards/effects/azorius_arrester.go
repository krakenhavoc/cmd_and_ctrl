package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azorius Arrester — Creature — Human Soldier {1}{W}, 2/1 (Ravnica: Clue
// Edition):
//
//	"When this creature enters, detain target creature an opponent
//	 controls. (Until your next turn, that creature can't attack or block
//	 and its activated abilities can't be activated.)"
//
// The target is chosen as the trigger goes on the stack; the detain is the
// Detain primitive, pinned to that creature as the trigger resolves (CR
// 701.35a) and ending as the controller's next turn begins.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f2a27a1-f9cf-4dac-a248-5c6af4de3722",
		Name:         "Azorius Arrester",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Azorius Arrester — detain target creature an opponent controls",
				detainEachLegalTarget("Azorius Arrester — detain")),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
