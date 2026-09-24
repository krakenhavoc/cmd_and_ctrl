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
// One clause does not ship, and it is declared rather than dropped
// silently (#259 — a card that quietly loses a clause is stronger than
// printed, never the direction to err):
//
//   - "Hexproof from monocolored" is protection-FAMILY vocabulary —
//     it tests the QUALITY of the source of a spell or ability, not
//     a simple targeting bit — and ADR 0038 §7 rules the whole family
//     out for exactly that reason: the targeting choke point never
//     receives the source's colour count to test against. Plain
//     hexproof (from everything) would be a real and strictly
//     stronger ability than what is printed, so it is left off
//     entirely rather than substituted.
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
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Hexproof from monocolored creatures and spells isn't implemented — the equipped creature gets no hexproof from this card.",
		},
		Static:    []game.StaticAbility{PumpAttached(2, 2)},
		Activated: []ActivatedAbility{equip},
	})
}
