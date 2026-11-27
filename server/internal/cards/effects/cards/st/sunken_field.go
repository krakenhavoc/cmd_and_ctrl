package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunken Field — Enchantment — Aura {1}{U} (Prophecy, S82 OLD b669, #5617):
//
//	"Enchant land
//	 Enchanted land has '{T}: Counter target spell unless its controller pays
//	 {1}.'"
//
// A granted activated ability on a land (Squirrel Nest's shape, ADR 0093)
// with Force Spike's pay-or-be-countered body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46569712-1f28-41cb-81be-fd938408c71f",
		Name:         "Sunken Field",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: "sunken-field/counter",
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Counter target spell unless its controller pays {1}.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.TapCost(),
				Targets: effects.TargetSpell("target spell"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					return effects.CounterUnlessPaid{
						StackID:  item.Targets[0].ID,
						Cost:     "{1}",
						Question: "Sunken Field — pay {1} or your spell is countered",
					}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{T}: Counter target spell unless its controller pays {1}.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached("sunken-field/counter")},
	})
}
