package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Defense Grid — Artifact {2} (Urza's Legacy, #5623):
//
//	"Each spell costs {3} more to cast except during its controller's
//	 turn."
//
// The Commander staple that taxes everything cast off-turn. It is a cost
// increase the cast reads as it totals the spell's cost (CR 601.2f), so a
// spell its own controller casts during their own turn pays nothing extra,
// and every spell cast during someone else's turn, Defense Grid's
// controller's included, costs {3} more. "Its controller" is the spell's
// caster, never the Grid's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "22ba474e-f6e6-450d-a4cf-0739f0b51079",
		Name:         "Defense Grid",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsMore(3, "Each spell costs {3} more to cast except during its controller's turn.", CastDuringAnotherPlayersTurn()),
		},
	})
}
