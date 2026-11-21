package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dawnhart Geist — Creature — Spirit Warlock {1}{W}, 1/3 (Innistrad:
// Crimson Vow, #5495):
//
//	"Whenever you cast an enchantment spell, you gain 2 life."
//
// Argothian Enchantress's cast condition over a life gain. It fires on
// casting, so a countered enchantment still gains the life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed7a2ee0-289d-4e5f-92e0-1539395baf33",
		Name:         "Dawnhart Geist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.EnchantmentSpellCastByYou, "Dawnhart Geist — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
