package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Song of Serenity — Enchantment {1}{G} (Exodus, S82 OLD batch 743, #5626):
//
//	"Creatures that are enchanted can't attack or block."
//
// A layer-6 restriction on every creature, anyone's, that has an Aura attached
// to it as the layers are applied (CantAttack and CantBlock). Equipment and
// other non-Aura attachments don't count; the restriction ends the moment the
// Aura does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c543799-1b72-4f32-a5be-475c0134ac0f",
		Name:         "Song of Serenity",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Creatures that are enchanted can't attack or block.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, _ *game.Card) bool {
				return effects.S82OldB743EnchantedCreature(target, g)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
	})
}
