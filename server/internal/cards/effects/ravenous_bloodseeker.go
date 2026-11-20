package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ravenous Bloodseeker — Creature — Vampire Berserker {1}{R}, 1/3
// (Shadows over Innistrad, #5543):
//
//	"Discard a card: This creature gets +2/-2 until end of turn."
//
// Noose Constrictor's discard cost with Breathstealer's +X/-X self-pump
// (thisGetsUntilEndOfTurn); a toughness of 0 dies to the state-based check
// (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc548ecb-57c4-49ef-a7aa-312bf3665fab",
		Name:         "Ravenous Bloodseeker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Discard a card: This creature gets +2/-2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: -2}},
			Cost:    DiscardN(1, "a card"),
			Effect:  thisGetsUntilEndOfTurn(2, -2, "Ravenous Bloodseeker — +2/-2 until end of turn"),
		}},
	})
}
