package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Colossus Hammer — Artifact — Equipment for {1} (EDHREC rank 824):
//
//	"Equipped creature gets +10/+10 and loses flying.
//	 Equip {8}"
//
// The first Equipment in the catalog whose static TAKES something
// away, and the reason RemoveFromAttached exists beside
// GrantToAttached. Both are layer 6 and both are scoped by the same
// attachment predicate; the removal simply runs the grant backwards.
//
// The drawback is real here and the engine honours it: flying is read
// by Game.CanBlockLocked, so a Hammered flier really can be blocked by anything
// on the ground. That is the whole cost of a one-mana +10/+10 — the
// equip cost being eight is the other, and the format's answer to it
// (Sigarda's Aid, Puresteel Paladin, Ardenn) is not in the catalog
// yet, so here the Hammer is honestly an eight-mana equip.
//
// One CR 613.7 consequence worth stating, because it is correct and
// surprising: the removal strips what the layers have granted so far
// in TIMESTAMP order. A flying grant from a source attached after the
// Hammer — Lightning Greaves is not one, but a later Aura could be —
// lands after this removal and survives it. That is the layer system
// working, not a leak, and it is the same limit b27LoseKeyword
// declares for Archetype of Aggression.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8ec03b88-8d3a-4a32-8b7c-7da59b0c03d0",
		Name:         "Colossus Hammer",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(10, 10),
			RemoveFromAttached("flying"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{8}"),
		},
	})
}
