package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood Sun — Enchantment {2}{R}:
//
//	"When this enchantment enters, draw a card.
//	 All lands lose all abilities except mana abilities."
//
// The static is LandsLoseAllAbilitiesExceptMana (#6246): every land on the
// battlefield, whoever controls it, loses its triggered, non-mana activated,
// static and replacement abilities and its keywords, and keeps its mana
// abilities (CR 613.1f, CR 605.1a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1534d90f-fc88-42dc-bf01-8b0986485e78",
		Name:         "Blood Sun",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Static:       []game.StaticAbility{effects.LandsLoseAllAbilitiesExceptMana()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Blood Sun — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
