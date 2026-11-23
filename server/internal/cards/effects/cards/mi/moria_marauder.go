package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moria Marauder — Creature — Goblin Warrior {R}{R}, 1/1 (The Lord of the
// Rings: Tales of Middle-earth Commander):
//
//	"Double strike
//	 Whenever a Goblin or Orc you control deals combat damage to a player,
//	 exile the top card of your library. You may play that card this turn."
//
// Kellan's impulse exile (b12ImpulseExileForTurn) on a Goblin-or-Orc combat
// damage watch. Double strike deals damage twice, so the Marauder itself
// triggers it twice. Each is one occurrence, and the tribes are read
// through the source's text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9f0027e3-815b-4965-9ef4-3c4814c62cf3",
		Name:            "Moria Marauder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.GapB126TribeDealtCombatDamageToAPlayer("Goblin", "Orc"),
				"Moria Marauder — exile the top card of your library; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					_, err := effects.B12ImpulseExileForTurn(g, item, 1)
					return err
				}),
		},
	})
}
