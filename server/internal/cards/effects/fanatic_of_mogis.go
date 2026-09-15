package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fanatic of Mogis — Creature — Minotaur Shaman {3}{R}, 4/2 (EDHREC
// rank 3810):
//
//	"When this creature enters, it deals damage to each opponent equal
//	 to your devotion to red. (Each {R} in the mana costs of
//	 permanents you control counts toward your devotion to red.)"
//
// Gray Merchant's red cousin: a burn spell whose size is the board.
// One ETB trigger; devotion is counted as it resolves (devotionTo —
// CR 700.5, hybrid symbols count for every colour they offer), the
// Fanatic's own {R} included while he is still on the battlefield.
// The damage is dealt by the Fanatic to each opponent, one event
// each, so a Fiendish Duo doubles it and a Sygg sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f19d06af-6caf-41d8-8d0a-d5d50bd67900",
		Name:         "Fanatic of Mogis",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fanatic of Mogis — damage to each opponent equal to your devotion to red", b36DamageEachOpponentForDevotionToRed),
		},
	})
}
