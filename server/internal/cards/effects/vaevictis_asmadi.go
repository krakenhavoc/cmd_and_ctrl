package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vaevictis Asmadi — Legendary Creature — Elder Dragon
// {2}{B}{B}{R}{R}{G}{G}, 7/7 (Legends, #5639):
//
//	"Flying
//	 At the beginning of your upkeep, sacrifice Vaevictis Asmadi unless
//	 you pay {B}{R}{G}.
//	 {B}: Vaevictis Asmadi gets +1/+0 until end of turn.
//	 {R}: Vaevictis Asmadi gets +1/+0 until end of turn.
//	 {G}: Vaevictis Asmadi gets +1/+0 until end of turn."
//
// The upkeep payment is UpkeepSacrificeUnlessYouPay. The three pumps are
// Shivan Dragon's, one row per colour, each the shared
// thisGetsUntilEndOfTurn body: no tap in the cost, so any number of
// activations, summoning sickness or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4d2cc2a8-e08a-420a-8922-c63e39129e23",
		Name:            "Vaevictis Asmadi",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			UpkeepSacrificeUnlessYouPay("Vaevictis Asmadi", "{B}{R}{G}"),
		},
		Activated: []ActivatedAbility{
			vaevictisAsmadiPump("{B}"),
			vaevictisAsmadiPump("{R}"),
			vaevictisAsmadiPump("{G}"),
		},
	})
}

// vaevictisAsmadiPump is one of the three "{C}: Vaevictis Asmadi gets
// +1/+0 until end of turn." rows.
func vaevictisAsmadiPump(cost string) ActivatedAbility {
	return ActivatedAbility{
		Label:   cost + ": Vaevictis Asmadi gets +1/+0 until end of turn.",
		Purpose: game.Purpose{Answers: game.AnswerPump},
		Cost:    ManaCost(cost),
		Effect:  thisGetsUntilEndOfTurn(1, 0, "Vaevictis Asmadi — +1/+0"),
	}
}
