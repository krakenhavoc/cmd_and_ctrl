package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampiric Rites — Enchantment {B}:
//
//	"{1}{B}, Sacrifice a creature: You gain 1 life and draw a card."
//
// A one-mana enchantment that turns every creature into a card. Like
// Bastion of Remembrance, the outlet itself survives creature
// removal, so it keeps working through the wipe that feeds it.
//
// The cost pairs mana with SACRIFICE-ANOTHER (SacrificeACreature),
// distinct from Mind Stone's sacrifice-self: the controller picks
// which creature from a list at announce, and the engine enforces
// CR 701.21a (it must be one they control).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "660de988-b6fb-4f36-8006-42af3e7f908d",
		Name:         "Vampiric Rites",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice a creature: You gain 1 life and draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Draws: 1},
			Cost: func() game.AbilityCost {
				c := effects.SacrificeACreature()
				c.Mana = "{1}{B}"
				return c
			}(),
			Effect: effects.Do(
				// The draw waits for a prompt the life gain asks (CR 608.2c, #4155).
				effects.GainLife{Amount: 1},
				effects.DrawCards{N: 1},
			),
		}},
	})
}
