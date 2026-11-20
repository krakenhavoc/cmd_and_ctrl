package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lifegift — Enchantment {2}{G} (Betrayers of Kamigawa, #5602):
//
//	"Whenever a land enters, you may gain 1 life."
//
// Any land, any player's: the condition is the land-entered read Zo-Zu
// the Punisher uses (b39LandEnteredController), one trigger per land. The
// "you may" is the trigger's own yes/no as it resolves (CR 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c99d4be7-244c-4fb5-877a-8aee961a2666",
		Name:         "Lifegift",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b39LandEnteredController(ev, g)
				return ok
			}, "Lifegift — you may gain 1 life", Do(GainLife{Amount: 1})),
				"Lifegift — gain 1 life?"),
		},
	})
}
