package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Professor Hulk — Legendary Creature — Gamma Scientist Hero {4}{U}{U},
// 6/6 (Marvel Super Heroes Commander, #4455):
//
//	"Trample
//	 Whenever Professor Hulk deals combat damage to a player, draw that
//	 many cards."
//
// Trample rides PrintedKeywords. The trigger is the shared "deals combat
// damage to a player" one; "that many" is the damage that event carries
// (trample's excess to the player counts, the part assigned to a blocker
// does not).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b8a627a1-d6d7-45d1-aeea-ff7b834cbb29",
		Name:            "Professor Hulk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Professor Hulk — draw that many cards",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return effects.DrawCards{Player: item.Controller, N: item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
