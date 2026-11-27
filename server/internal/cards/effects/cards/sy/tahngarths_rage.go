package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tahngarth's Rage — Enchantment — Aura {R} (Tempest, S82 OLD, #5628):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+0 as long as it's attacking. Otherwise, it gets
//	 -2/-1."
//
// Two layer 7c modifications read off the attachment and the creature's combat
// status as the layers are asked: +3/+0 while it is an attacking creature,
// -2/-1 the rest of the time (so a creature that is not attacking, blocking
// ones included, shrinks). Kor Scythemaster's attacking-status dependency makes
// the layer cache follow the declaration and the end of combat.
//
// No simplification.
func init() {
	attacking := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.AttachedToSource(target, g, source) && target.AttackingTarget != uuid.Nil
	}
	idle := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.AttachedToSource(target, g, source) && target.AttackingTarget == uuid.Nil
	}
	pump := func(label string, applies func(*game.Card, *game.Game, *game.Card) bool, power, toughness int) game.StaticAbility {
		return game.StaticAbility{
			Label:                    label,
			Layer:                    game.Layer7PT,
			SubLayer:                 game.SubLayer7C_Modify,
			AppliesTo:                applies,
			DependsOnAttackingStatus: true,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += power
				c.Toughness += toughness
			},
		}
	}
	effects.Register(effects.Spec{
		OracleID:     "5b0cb089-087d-4b91-9e63-8186dc737a00",
		Name:         "Tahngarth's Rage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			pump("Enchanted creature gets +3/+0 as long as it's attacking.", attacking, 3, 0),
			pump("Enchanted creature gets -2/-1 as long as it isn't attacking.", idle, -2, -1),
		},
	})
}
