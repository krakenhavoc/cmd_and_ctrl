package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hot Springs — Enchantment — Aura {1}{G} (Ice Age, S82 OLD b828, #5636):
//
//	"Enchant land you control
//	 Enchanted land has "{T}: Prevent the next 1 damage that would be dealt to
//	 any target this turn.""
//
// Find the Path's granted ability on the enchanted land, with Abuna Acolyte's
// prevention shield as its effect.
//
// No simplification.
func init() {
	const grant = "hot-springs/prevent"
	effects.Register(effects.Spec{
		OracleID:     "123880ad-7b54-4ab0-82eb-c78b12ab89a7",
		Name:         "Hot Springs",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Effect:  effects.PreventNextOneDamageToTheTarget("Hot Springs"),
			}},
			Text: "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
	})
}
