package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Darksteel Reactor — Artifact {4}:
//
//	"Indestructible
//	 At the beginning of your upkeep, you may put a charge counter on
//	 this artifact.
//	 When this artifact has twenty or more charge counters on it, you
//	 win the game."
//
// ADR 0107 §1 (#1858). Indestructible is the printed keyword. The upkeep
// counter is a CR 603.5 "you may". The win is a CR 603.8 state trigger:
// it triggers however the twentieth counter arrived (a proliferate
// counts), and "you" is the Reactor's controller as the ability resolves.
// A player who can't win (an opponent's Platinum Angel) doesn't, and the
// ability triggers again once it has left the stack (WinTheGame, ADR
// 0057).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc483bab-14fb-498d-9310-9c070766c7ae",
		Name:         "Darksteel Reactor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(AtYourUpkeep("Darksteel Reactor — put a charge counter", putACounterOnThis(game.CounterCharge)),
				"Darksteel Reactor — put a charge counter on it?"),
			WhenThisHasAtLeast(game.CounterCharge, 20, "Darksteel Reactor — you win the game", Do(WinTheGame{})),
		},
	})
}
