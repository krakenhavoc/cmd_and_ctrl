package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Molting Snakeskin — Enchantment — Aura {B} (Khans of Tarkir):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+0 and has "{2}{B}: Regenerate this creature.""
//
// Infernal Scarring's pump and granted ability, with River Boa's regeneration as
// the granted activated row: the creature, not the Aura, is the source, so its
// controller pays and makes the shield, and both are lost with the Aura.
//
// No simplification.
func init() {
	const grant = "molting-snakeskin/regenerate"
	const text = "{2}{B}: Regenerate this creature."
	effects.Register(effects.Spec{
		OracleID:     "1a668556-da98-4f1e-b1f7-26b7d63357a1",
		Name:         "Molting Snakeskin",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{2}{B}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{effects.PumpAttached(2, 0), effects.GrantAbilitiesToAttached(grant)},
	})
}
