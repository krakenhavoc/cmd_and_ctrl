package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merfolk Coralsmith — Creature — Merfolk {2}{U}, 2/3 (Wilds of Eldraine,
// #4471):
//
//	"{1}: This creature gets +1/-1 until end of turn.
//	 When this creature dies, scry 2."
//
// The pump is Unliving Psychopath's self boost for {1}, so a fourth
// activation kills it, and that death scries. The scry is its
// controller's, and needs nothing of the creature once it has died.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "05f3638f-3c66-47bb-8a82-ad173e152573",
		Name:         "Merfolk Coralsmith",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}: This creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}"),
			Effect:  thisGetsUntilEndOfTurn(1, -1, "Merfolk Coralsmith — +1/-1 until end of turn"),
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Merfolk Coralsmith — scry 2", Do(Scry{N: 2})),
		},
	})
}
