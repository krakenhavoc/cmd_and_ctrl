package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "39c10798-9dfc-47a2-9da3-b4038509aebe",
		Name:         "Underworld Connections",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []AbilityGrant{{
			Key: underworldConnectionsGrant,
			Activated: []ActivatedAbility{{
				Label:   "{T}, Pay 1 life: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    Plus(TapCost(), PayLife(1)),
				Effect:  Do(DrawCards{N: 1}),
			}},
			Text: "{T}, Pay 1 life: Draw a card.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(underworldConnectionsGrant)},
	})
}

const underworldConnectionsGrant = "underworld-connections/draw"
