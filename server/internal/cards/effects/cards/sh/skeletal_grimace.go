package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const skeletalGrimaceRegenerateGrant = "skeletal-grimace/regenerate"

// Skeletal Grimace — Enchantment — Aura {1}{B}:
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has "{B}: Regenerate this creature.""
//
// Trollhide's shape: the quoted ability is the enchanted creature's own, so it
// regenerates itself, and moving the Aura moves the ability.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "99e03fa0-8e0c-4b5d-880f-d7de201d5f41",
		Name:         "Skeletal Grimace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: skeletalGrimaceRegenerateGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{B}: Regenerate this creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{B}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{B}: Regenerate this creature.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantAbilitiesToAttached(skeletalGrimaceRegenerateGrant),
		},
	})
}
