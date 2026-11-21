package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vault Robber — Creature — Dwarf Rogue {1}{R}, 2/1 (Jumpstart 2022, #5948):
//
//	"{1}, {T}, Exile a creature card from your graveyard: Create a Treasure
//	 token."
//
// Fungal Plots' graveyard-exile cost: the exile is a cost (CR 118.8), paid
// as the ability is activated, so an empty graveyard refuses the activation.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68c90a32-d6ee-4595-8517-5f4cb0e4062d",
		Name:         "Vault Robber",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Exile a creature card from your graveyard: Create a Treasure token.",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.ExileFromGraveyard(1, "a creature card", effects.IsCreatureCard)),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1}),
		}},
	})
}
