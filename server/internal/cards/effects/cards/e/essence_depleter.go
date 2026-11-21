package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Essence Depleter — Creature — Eldrazi Drone {2}{B}, 2/3 (Oath of the Gatewatch, #5545):
//
//	"Devoid (This card has no color.)
//	 {1}{C}: Target opponent loses 1 life and you gain 1 life. ({C}
//	 represents colorless mana.)"
//
// Devoid makes it colourless (CR 702.114a). The drain is Blood Hustler's
// row. {C} is colorless mana only, never generic (CR 107.4c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "19ea37cc-e715-4a3c-87f3-5ed9d44f42d9",
		Name:            "Essence Depleter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{C}: Target opponent loses 1 life and you gain 1 life.",
			Cost:    effects.ManaCost("{1}{C}"),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect:  effects.DrainTargetOpponentOne,
		}},
	})
}
