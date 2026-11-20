package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Devkarin Dissident — Creature — Elf Warrior {1}{G}, 2/2 (Guilds of Ravnica, #5525):
//
//	"{4}{G}: This creature gets +2/+2 until end of turn."
//
// Battlefield Percher's pump at {4}{G} for +2/+2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a671fedf-3766-4fda-abeb-6f63e2fdd20b",
		Name:         "Devkarin Dissident",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}{G}: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost:    ManaCost("{4}{G}"),
			Effect:  thisGetsUntilEndOfTurn(2, 2, "Devkarin Dissident — +2/+2 until end of turn"),
		}},
	})
}
