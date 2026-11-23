package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mishra's Domination — Enchantment — Aura {1}{R} (The Brothers' War):
//
//	"Enchant creature
//	 As long as you control enchanted creature, it gets +2/+2. Otherwise,
//	 it can't block."
//
// "You" is the Aura's controller; the condition is read live, so a creature
// that changes hands flips from one half to the other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ede2354a-64b0-402c-8fbe-9da90f811a9b",
		Name:         "Mishra's Domination",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(func(host *game.Card, g *game.Game, source *game.Card) bool {
				return effects.AttachedToSource(host, g, source) && host.Controller == source.Controller
			}, 2, 2),
			effects.RestrictAttachedWhile(func(host *game.Card, _ *game.Game, source *game.Card) bool {
				return host.Controller != source.Controller
			}, game.CantBlock),
		},
	})
}
