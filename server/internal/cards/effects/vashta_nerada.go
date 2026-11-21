package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vashta Nerada — Creature — Alien Horror {2}{B}, 1/1 (Doctor Who, #5948):
//
//	"Indestructible
//	 Shadow
//	 Morbid — At the beginning of each end step, if a creature died this turn,
//	 put a +1/+1 counter on this creature."
//
// Bulette's end-step trigger on EACH end step rather than yours: the
// intervening "if" (CondAnyCreatureDied, any controller's creature, tokens
// included) is checked as the step begins and again on resolution (CR 603.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9ae26054-caf4-47a3-b4bb-3dc14a634740",
		Name:            "Vashta Nerada",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible", "shadow"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondAnyCreatureDied{},
				On(game.EventBeginEndStep, s82AfrB07AnyCreatureDiedWhen,
					"Vashta Nerada — put a +1/+1 counter on this creature",
					Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}
