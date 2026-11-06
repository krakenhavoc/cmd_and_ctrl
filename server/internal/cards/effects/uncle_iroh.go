package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8a98ddac-7341-4118-b188-51d8d03afd1d",
		Name:         "Uncle Iroh",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Uncle Iroh — firebending 1", Do(firebending(1))),
		},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Lesson spells you cast cost {1} less to cast.", YourSpell(), avB18LessonSpell()),
		},
	})
}
