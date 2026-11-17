package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unspeakable Symbol — Enchantment {1}{B}{B} (Scourge):
//
//	"Pay 3 life: Put a +1/+1 counter on target creature."
//
// An activated ability whose whole cost is 3 life (Staff of Compleation's
// PayLife), with no tap and no limit, so it can be activated any number of
// times while its controller has the life (CR 119.4). The counter goes on
// any creature, any player's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a5690d5f-633c-4a1e-afba-5fd79dcbf20e",
		Name:         "Unspeakable Symbol",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Pay 3 life: Put a +1/+1 counter on target creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    PayLife(3),
			Targets: TargetCreature("target creature"),
			Effect:  putPlusOneCounterOnEachLegalTarget,
		}},
	})
}
