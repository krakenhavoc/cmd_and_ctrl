package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vorrac Battlehorns — Artifact — Equipment, {2}:
//
//	"Equipped creature has trample and can't be blocked by more than
//	 one creature.
//	 Equip {1}"
//
// The reference card for a COUNT rule on an attached creature (#750,
// ADR 0045 addendum Decision 12): MaxBlockers(OnAttached(), 1). A
// block of two is refused whole at declaration with
// too_many_blockers, and nothing is stored. On a menace creature the
// minimum of two exceeds the maximum of one, so the creature can't be
// blocked at all — the printed interaction, and the engine's.
//
// Trample is the ordinary layer-6 grant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "89750d72-d1c0-4c8a-aa72-fdd48570aa92",
		Name:         "Vorrac Battlehorns",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("trample"),
		},
		BlockRules: []game.BlockRule{
			MaxBlockers(OnAttached(), 1),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}
