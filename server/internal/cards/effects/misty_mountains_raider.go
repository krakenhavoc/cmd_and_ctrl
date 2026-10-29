package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Misty Mountains Raider — Creature — Goblin Soldier {4}{R}, 4/4
// (The Hobbit, HOB):
//
//	"Whenever you attack, amass Goblins 2."
//
// "Whenever you attack" is one trigger per attack declaration, not one per
// attacker (OncePerBatch, ADR 0129 §3): the engine emits EventAttack per
// creature, and the batch declines every later event of the same declaration.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c394b767-620c-4310-b78f-ece312259e46",
		Name:         "Misty Mountains Raider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Misty Mountains Raider — amass Goblins 2", Do(Amass{Subtype: "Goblin", N: 2}))),
		},
	})
}
