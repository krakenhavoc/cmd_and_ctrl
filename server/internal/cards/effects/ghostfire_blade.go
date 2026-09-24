package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghostfire Blade — Artifact — Equipment for {1}:
//
//	"Equipped creature gets +2/+2.
//	 Equip {3}
//	 This Equipment's equip ability costs {2} less to activate if it
//	 targets a colorless creature."
//
// A proof card for the ability's own cost clause (#1296,
// ActivatedAbility.CostModifiers): the discount is a condition on the
// equip's TARGET, judged when the cost is determined — after the
// target is announced (CR 602.2b → 601.2c, then 601.2f) — so an
// artifact creature or a face-down 2/2 is equipped for {1} and a
// coloured one for the printed {3}.
//
// The clause is printed as a separate static line rather than inside
// the equip ability, but it names exactly one ability ("this
// Equipment's equip ability"), which is what the ability's own slot
// expresses; a board modifier would also price every OTHER equip
// ability of every other permanent the controller has.
func init() {
	equip := EquipAbility("{3}")
	equip.CostModifiers = []game.CostModifier{
		CostsLessIfItTargets(2,
			"This Equipment's equip ability costs {2} less to activate if it targets a colorless creature.",
			Colorless()),
	}
	Register(Spec{
		OracleID:     "093b1d8d-c836-4e9b-855b-4020361aa6ac",
		Name:         "Ghostfire Blade",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 2)},
		Activated:    []ActivatedAbility{equip},
	})
}
