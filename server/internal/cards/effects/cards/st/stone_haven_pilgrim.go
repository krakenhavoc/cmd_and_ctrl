package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stone Haven Pilgrim — Creature — Kor Cleric {1}{W}, 2/2 (Jumpstart,
// #5510):
//
//	"Whenever this creature attacks, if you control an artifact or
//	 enchantment, this creature gets +1/+1 and gains lifelink until end of
//	 turn."
//
// The "if" is intervening (CR 603.4): read as the creature attacks and again
// as the trigger resolves, so losing the last artifact or enchantment in
// response leaves it doing nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "264fafb7-3bd3-4bf1-81b2-2e436c9ca8d0",
		Name:         "Stone Haven Pilgrim",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control an artifact or enchantment",
				effects.YouControlAtLeast(1, func(c game.Card) bool { return c.IsArtifact() || c.IsEnchantment() }),
				effects.On(game.EventAttack, effects.ThisAttacked,
					"Stone Haven Pilgrim — it gets +1/+1 and gains lifelink until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						return effects.SocB27ThisGetsAndGains(effects.NewContext(g, item), 1, "lifelink", "Stone Haven Pilgrim")
					})),
		},
	})
}
