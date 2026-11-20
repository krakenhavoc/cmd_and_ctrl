package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hero of the Pride — Creature — Cat Soldier {1}{W}, 2/2 (Theros Beyond Death, #5514):
//
//	"Whenever you cast a spell that targets this creature, creatures
//	 you control get +1/+0 until end of turn."
//
// Heroic is Hero of the Games' trigger: a spell its controller casts
// that targets it pumps every creature they control.
//
// No simplification.
func init() {
	const label = "Hero of the Pride — creatures you control get +1/+0 until end of turn (heroic)"
	Register(Spec{
		OracleID:     "f3a2d80c-de6f-4b0e-aeb1-78639d55a6e6",
		Name:         "Hero of the Pride",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			heroic(label, func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Label: label}.Apply(NewContext(g, item))
			}),
		},
	})
}
