package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valakut Fireboar — Creature — Elemental Boar {4}{R}, 1/7:
//
//	"Whenever this creature attacks, switch its power and toughness until
//	 end of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
// The ability's source must still be the same object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "126bf98a-64d0-4de2-9bf4-f59d36ef6e25",
		Name:         "Valakut Fireboar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Valakut Fireboar — switch its power and toughness", switchThisPT("Valakut Fireboar — switch its power and toughness")),
		},
	})
}
