package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kozilek's Translator — {4}{B} Creature — Eldrazi Drone 3/5 (Oath of the
// Gatewatch, S82 OLD b295):
//
//	"Devoid (This card has no color.)
//	 Pay 1 life: Add {C}. Activate only once each turn."
//
// A mana ability (CR 605.1a) whose cost is one life and which may be used once
// each turn (Barrels of Blasting Jelly's limit); no tap, so a Translator that
// just arrived can use it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6766c5e4-baa7-4db4-a900-c08fbcd98f5d",
		Name:            "Kozilek's Translator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		ManaAbilities: []effects.ManaAbility{{
			Cost:        effects.ManaAbilityCost{Life: 1},
			Produced:    "{C}",
			Label:       "Pay 1 life: Add {C}. Activate only once each turn.",
			OncePerTurn: true,
		}},
	})
}
