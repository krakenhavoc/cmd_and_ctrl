package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Veiled Sentry — Enchantment {U}:
//
//	"When an opponent casts a spell, if this permanent is an enchantment, it
//	 becomes an Illusion creature with power and toughness each equal to that
//	 spell's mana value."
//
// The size is the casting spell's mana value, read as the ability resolves
// (from the spell's last known information if it has been countered or has
// resolved), and it is a plain number set once, not a characteristic-defining
// ability that tracks anything afterwards. An intervening "if" (CR 603.4).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b9b8e60-51eb-4d4a-95c1-48ca5b8153df",
		Name:         "Veiled Sentry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.OldB717AnOpponentCastIfEnchantment(nil),
				"Veiled Sentry — it becomes an Illusion creature with power and toughness each equal to that spell's mana value",
				effects.OldB720BecomesACreatureSized([]string{"Illusion"}, func(g *game.Game, item *game.StackItem) (int, int) {
					n := effects.TriggeringSpellManaValue(g, item)
					return n, n
				}, "Veiled Sentry — an Illusion creature")),
		},
	})
}
