package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wakandan Shield Guard — Creature — Human Soldier {1}{W}, 1/1 (Marvel
// Super Heroes Commander, #4455): "When this creature enters, create a
// 1/1 white Soldier creature token." A one-token enters trigger from the
// catalog's token table.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a9268e61-2697-46ff-b219-4f2be646cbe4",
		Name:         "Wakandan Shield Guard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wakandan Shield Guard — create a 1/1 white Soldier creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Soldier"), N: 1})),
		},
	})
}
