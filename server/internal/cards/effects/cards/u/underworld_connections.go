package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Underworld Connections — Enchantment — Aura {1}{B}{B}:
//
//	"Enchant land
//	 Enchanted land has "{T}, Pay 1 life: Draw a card.""
//
// A granted activated ability on a land (Squirrel Nest, ADR 0093): the
// land's controller pays the tap and the life and draws, which is not
// always the Aura's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "39c10798-9dfc-47a2-9da3-b4038509aebe",
		Name:         "Underworld Connections",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: underworldConnectionsGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}, Pay 1 life: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    effects.Plus(effects.TapCost(), effects.PayLife(1)),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			}},
			Text: "{T}, Pay 1 life: Draw a card.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(underworldConnectionsGrant)},
	})
}

const underworldConnectionsGrant = "underworld-connections/draw"
