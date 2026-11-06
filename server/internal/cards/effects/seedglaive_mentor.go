package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seedglaive Mentor — Creature — Mouse Soldier {1}{R}{W}, 3/2 (Bloomburrow,
// #5230):
//
//	"Vigilance, haste
//	 Valiant — Whenever this creature becomes the target of a spell or
//	 ability you control for the first time each turn, put a +1/+1 counter
//	 on it."
//
// The printed keywords, and the shared valiant trigger (valiant.go) with
// Heartfire Hero's counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f50bbe5e-a839-40e1-ba79-1e96f91bd218",
		Name:            "Seedglaive Mentor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "haste"},
		Triggered: []game.TriggeredAbility{
			Valiant("Seedglaive Mentor — Valiant: put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
