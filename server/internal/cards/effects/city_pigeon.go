package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// City Pigeon — Creature — Bird {W}, 1/1 (Marvel Super Heroes Commander,
// #4455): "Flying. When this creature leaves the battlefield, create a
// Food token." A leaves-the-battlefield trigger (CR 603.10a) that makes
// the catalog Food token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ac8cae63-270c-4f73-b29b-50f8f2395fd9",
		Name:            "City Pigeon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisLeaves("City Pigeon — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
	})
}
