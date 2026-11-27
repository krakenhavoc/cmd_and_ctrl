package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Presence of the Master — Enchantment {3}{W} (Urza's Saga, S82 OLD b721):
//
//	"Whenever a player casts an enchantment spell, counter it."
//
// Nullstone Gargoyle's "counter that spell" on the cast trigger, for any
// player's enchantment spell, its controller's included. A spell that can't
// be countered stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1aad679-93ec-420e-881b-35ebc99763a2",
		Name:         "Presence of the Master",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.TmcB8APlayerCast(effects.Enchantment()),
				"Presence of the Master — counter that enchantment spell", effects.S82OldB582CounterThatSpell),
		},
	})
}
