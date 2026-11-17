package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flowstone Charger — Creature — Beast {2}{R}{W}, 2/5 (Apocalypse,
// #5614):
//
//	"Whenever this creature attacks, it gets +3/-3 until end of turn."
//
// Soltari Trooper's attack trigger (WheneverThisAttacks, CR 508.1m)
// with a +3/-3 thisGetsUntilEndOfTurn body, so the printed 2/5 attacks
// as a 5/2. The change ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a19e0a7f-d356-4d2c-8414-70a40ce18674",
		Name:         "Flowstone Charger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Flowstone Charger — it gets +3/-3 until end of turn",
				thisGetsUntilEndOfTurn(3, -3, "Flowstone Charger — +3/-3")),
		},
	})
}
