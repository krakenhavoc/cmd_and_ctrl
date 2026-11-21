package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uncle Iroh — Legendary Creature — Human Noble Ally {1}{R/G}{R/G}, 4/2
// (Avatar: The Last Airbender, #4459):
//
//	"Firebending 1 (Whenever this creature attacks, add {R}. This mana
//	 lasts until end of combat.)
//	 Lesson spells you cast cost {1} less to cast."
//
// Firebending is Avatar Roku's (`firebending`). The reduction is Bureau
// Headmaster's subtype shape for Lesson spells; a reduction applies to
// the generic part of the total cost and never changes the spell's mana
// value (the ruling, CR 601.2f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a98ddac-7341-4118-b188-51d8d03afd1d",
		Name:         "Uncle Iroh",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Uncle Iroh — firebending 1", effects.Do(effects.NewFirebending(1))),
		},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Lesson spells you cast cost {1} less to cast.", effects.YourSpell(), effects.AvB18LessonSpell()),
		},
	})
}
