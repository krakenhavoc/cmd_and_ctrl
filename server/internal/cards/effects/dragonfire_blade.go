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
// Two clauses do not ship, and both are declared rather than dropped
// silently (#259 — a card that quietly loses a clause is stronger
// than printed, never the direction to err):
//
//   - "Hexproof from monocolored" is protection-FAMILY vocabulary —
//     it tests the QUALITY of the source of a spell or ability, not
//     a simple targeting bit — and ADR 0038 §7 rules the whole family
//     out for exactly that reason: the targeting choke point never
//     receives the source's colour count to test against. Plain
//     hexproof (from everything) would be a real and strictly
//     stronger ability than what is printed, so it is left off
//     entirely rather than substituted.
//   - The equip cost discount is a cost REDUCTION on an ACTIVATED
//     ability keyed to the chosen target's colour count.
//     docs/engine-seams.md lists ability-cost modification as an open
//     seam (no component of game.AbilityCost carries a discount), so
//     equip always costs its full printed {4} — weaker than printed,
//     never stronger.
func init() {
	Register(Spec{
		OracleID:     "e809b847-b712-4558-95ea-9bb7356cde91",
		Name:         "Dragonfire Blade",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Hexproof from monocolored creatures and spells isn't implemented — the equipped creature gets no hexproof from this card.",
			"Equip always costs its full four mana — it never gets cheaper for the colors of the creature you're attaching it to.",
		},
		Static: []game.StaticAbility{PumpAttached(2, 2)},
		Activated: []ActivatedAbility{
			EquipAbility("{4}"),
		},
	})
}
