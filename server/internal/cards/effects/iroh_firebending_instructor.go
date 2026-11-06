package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iroh, Firebending Instructor — Legendary Creature — Human Noble Ally
// {2}{R}, 2/2 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Whenever Iroh attacks, attacking creatures get +1/+1 until end of
//	 turn."
//
// The set is read once as the trigger resolves (CR 611.2c): a creature
// put onto the battlefield attacking later gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e03797d-dd61-439f-b4ce-996cd71bdcf5",
		Name:         "Iroh, Firebending Instructor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Iroh, Firebending Instructor — attacking creatures get +1/+1 until end of turn",
				Do(BoostUntilEOT{Match: AttackingCreature(), Power: 1, Toughness: 1, Label: "Iroh, Firebending Instructor"})),
		},
	})
}
