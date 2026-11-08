package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Carrot Cake — Artifact — Food {1}{W} (Bloomburrow, #4467):
//
//	"When this artifact enters and when you sacrifice it, create a 1/1
//	 white Rabbit creature token and scry 1.
//	 {2}, {T}, Sacrifice this artifact: You gain 3 life."
//
// Heaped Harvest's shape: one printed ability with two conditions, so one
// declaration watches the Cake's own entry and its own sacrifice
// (b26SelfEnteredOrWasSacrificed). EventSacrifice fires while the Cake is
// still on the battlefield, and it fires for any sacrifice, the Food
// ability's own cost or a forage, as the ruling says. Only a permanent's
// controller can sacrifice it (CR 701.21a), so "you" needs no check. The
// Rabbit is created first, then the scry. The Food ability is the
// ordinary one (samiFoodLifeAbility).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5fe3aabd-3271-41bf-ae50-f6efaba7c4ae",
		Name:         "Carrot Cake",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1}, // the enters half of the Rabbit trigger
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventSacrifice}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b26SelfEnteredOrWasSacrificed(ev, source)
			}, "Carrot Cake — create a 1/1 white Rabbit and scry 1",
				Do(CreateToken{Template: TokenCard("1/1 white Rabbit"), N: 1}, Scry{N: 1})),
		},
		Activated: []ActivatedAbility{samiFoodLifeAbility()},
	})
}
