package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragonfire Blade — Artifact — Equipment for {1}:
//
//	"Equipped creature gets +2/+2 and has hexproof from monocolored.
//	 Equip {4}. This ability costs {1} less to activate for each color
//	 of the creature it targets."
//
// The pump and the plain Equip {4} are the ordinary shapes.
//
// The discount is the equip ability's OWN cost clause
// (ActivatedAbility.CostModifiers, #1296), read off the target the
// activation announces: CR 602.2b runs 601.2c (targets) before 601.2f
// (the total cost), so the colours counted are the target's as they
// stand when the cost is determined — a Vivi Ornitier ({U}{R}) is
// equipped for {2}, a five-colour commander for {0}, an artifact
// creature for the full {4}. The generic floor is the engine's:
// nothing here can take the cost below zero.
//
// "Hexproof from monocolored" is a layer-6 grant of
// game.HexproofFromMonocolored (#3172, ADR 0038's amendment of
// 2026-10-13): an opponent's source with exactly one colour can't
// target the equipped creature, a multicolored or colourless one can.
// It is not plain hexproof, which would be strictly stronger.
func init() {
	equip := EquipAbility("{4}")
	equip.CostModifiers = []game.CostModifier{
		CostsLessForTheCardItTargets(
			"This ability costs {1} less to activate for each color of the creature it targets.",
			ColorsOf),
	}
	Register(Spec{
		OracleID:     "e809b847-b712-4558-95ea-9bb7356cde91",
		Name:         "Dragonfire Blade",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 2), GrantToAttached(game.HexproofFromMonocolored)},
		Activated:    []ActivatedAbility{equip},
	})
}
