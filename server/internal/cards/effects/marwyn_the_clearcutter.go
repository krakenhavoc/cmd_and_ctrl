package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marwyn, the Clearcutter — Legendary Creature — Elf Warrior {R}, 2/1:
//
//	"{2}, {T}, Sacrifice an artifact or land: Draw a card."
//
// The sacrifice is a cost, paid at announce, and Marwyn herself is not an
// artifact or a land, so she is never a legal pick. The tap makes it
// subject to summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aff4e43d-cafa-47de-8e39-011f9254a26a",
		Name:         "Marwyn, the Clearcutter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice an artifact or land: Draw a card",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost: Plus(ManaCost("{2}"), TapCost(),
				game.AbilityCost{SacrificeOther: sacrificeSpec("an artifact or land", Or(Artifact(), Land()))}),
			Effect: Do(DrawCards{N: 1}),
		}},
	})
}
