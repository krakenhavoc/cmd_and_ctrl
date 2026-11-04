package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Curiosity Crafter — Creature — Bird Wizard {3}{U}, 3/3:
//
//	"Flying
//	 You have no maximum hand size.
//	 Whenever a creature token you control deals combat damage to a
//	 player, draw a card."
//
// "No maximum hand size" is Spec.NoMaxHandSize, which the engine derives
// at cleanup from the battlefield (Thought Vessel's bit). The draw is one
// trigger per token and damage event, so three tokens connecting draw
// three cards; a nontoken creature and an opponent's token draw nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0be907b2-54f9-4c8c-b6af-7c0093892e42",
		Name:            "Curiosity Crafter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		NoMaxHandSize:   true,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, socB3CreatureTokenYouControlDealtCombatDamageToAPlayer,
				"Curiosity Crafter — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
