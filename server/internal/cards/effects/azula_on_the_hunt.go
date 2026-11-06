package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azula, On the Hunt — Legendary Creature — Human Noble {3}{B}, 4/3:
//
//	"Firebending 2 (Whenever this creature attacks, add {R}{R}. This mana
//	 lasts until end of combat.)
//	 Whenever Azula attacks, you lose 1 life and create a Clue token."
//
// Two attack triggers, each its own row. Firebending is Avatar Roku's
// (`firebending`, #3172): the two red last until end of combat and the
// trigger uses the stack (the rulings). The second is Bitterblossom's
// lose-then-create on an attack; creating the Clue is not investigating
// (the ruling), so it is a plain token creation.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "89f96032-5cba-404d-82b0-38d2be963e4b",
		Name:         "Azula, On the Hunt",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Azula, On the Hunt — firebending 2", Do(firebending(2))),
			WheneverThisAttacks("Azula, On the Hunt — you lose 1 life and create a Clue token", Do(
				// The Clue waits for a prompt the life loss asks (CR 608.2c, #4155).
				GainLife{Amount: -1},
				CreateToken{Template: ClueToken(), N: 1},
			)),
		},
	})
}
