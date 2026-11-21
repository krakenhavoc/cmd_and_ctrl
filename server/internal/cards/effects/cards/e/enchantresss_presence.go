package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enchantress's Presence — Enchantment, {2}{G} (EDHREC rank 985):
//
//	"Whenever you cast an enchantment spell, draw a card."
//
// The original enchantress, on an enchantment rather than a body so
// it dodges creature removal. Mesa Enchantress's trigger without the
// "may" — mandatory, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "795b096a-2bce-4588-a2c9-abc5ea40dc0c",
		Name:         "Enchantress's Presence",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.EnchantmentSpellCastByYou, "Enchantress's Presence — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
