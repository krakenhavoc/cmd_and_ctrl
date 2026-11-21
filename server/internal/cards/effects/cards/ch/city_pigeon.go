package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// City Pigeon — Creature — Bird {W}, 1/1 (Marvel Super Heroes Commander,
// #4455): "Flying. When this creature leaves the battlefield, create a
// Food token." A leaves-the-battlefield trigger (CR 603.10a) that makes
// the catalog Food token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ac8cae63-270c-4f73-b29b-50f8f2395fd9",
		Name:            "City Pigeon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisLeaves("City Pigeon — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
	})
}
