package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cultivator Drone — {2}{U} Creature — Eldrazi Drone 2/3 (Oath of the
// Gatewatch, S82 OLD b296):
//
//	"Devoid (This card has no color.)
//	 {T}: Add {C}. Spend this mana only to cast a colorless spell, activate an
//	 ability of a colorless permanent, or pay a cost that contains {C}."
//
// A tap mana ability whose mana carries the #6259 restriction tag: an OR of a
// cast of a colorless spell, an activation of a colorless permanent's ability,
// and any cost that prints {C} (the whole cost, generic included).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6e62d553-d0d0-400e-a423-f0c3c1fdc911",
		Name:            "Cultivator Drone",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     "{C}",
			Label:        "{T}: Add {C}. Spend this mana only to cast a colorless spell, activate an ability of a colorless permanent, or pay a cost that contains {C}.",
			Restrictions: []string{effects.OldB296ColorlessManaRestriction()},
		}},
	})
}
