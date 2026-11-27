package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flowering Field — Enchantment — Aura {1}{W} (Prophecy, S82 OLD b669, #5617):
//
//	"Enchant land
//	 Enchanted land has '{T}: Prevent the next 1 damage that would be dealt to
//	 any target this turn.'"
//
// A granted activated ability on a land (Squirrel Nest's shape, ADR 0093):
// the land's controller activates it, with Benevolent Ancestor's one-point
// shield as its body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "40155cfb-078a-4359-9551-9c71266d8a1c",
		Name:         "Flowering Field",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: "flowering-field/prevent",
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Effect:  effects.PreventNextOneDamageToTheTarget("Flowering Field"),
			}},
			Text: "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached("flowering-field/prevent")},
	})
}
