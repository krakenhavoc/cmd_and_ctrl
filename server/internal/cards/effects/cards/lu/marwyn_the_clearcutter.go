package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "aff4e43d-cafa-47de-8e39-011f9254a26a",
		Name:         "Marwyn, the Clearcutter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice an artifact or land: Draw a card",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost: effects.Plus(effects.ManaCost("{2}"), effects.TapCost(),
				game.AbilityCost{SacrificeOther: effects.SacrificeSpec("an artifact or land", effects.Or(effects.Artifact(), effects.Land()))}),
			Effect: effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
