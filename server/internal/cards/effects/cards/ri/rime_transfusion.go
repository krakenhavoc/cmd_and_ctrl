package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rime Transfusion — Snow Enchantment — Aura {1}{B} (Coldsnap, #5597):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+1 and has "{S}: This creature can't be blocked
//	 this turn except by snow creatures.""
//
// Molting Snakeskin's pump and granted ability; the granted ability's source is the
// enchanted creature, which gets "can't be blocked this turn except by snow creatures"
// (Gingerbrute's evasion, over the snow supertype). {S} is snow mana.
//
// No simplification.
func init() {
	const grant = "rime-transfusion/snow-evasion"
	const text = "{S}: This creature can't be blocked this turn except by snow creatures."
	effects.Register(effects.Spec{
		OracleID:     "d7ff0525-8617-4e0f-92db-660f67d706ca",
		Name:         "Rime Transfusion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{S}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CantBeBlockedThisTurnExceptBy{
						Target:     item.SourceCardID,
						Supertypes: []string{"Snow"},
						Text:       "snow creatures",
						Label:      "Rime Transfusion — can't be blocked except by snow creatures",
					}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{effects.PumpAttached(2, 1), effects.GrantAbilitiesToAttached(grant)},
	})
}
