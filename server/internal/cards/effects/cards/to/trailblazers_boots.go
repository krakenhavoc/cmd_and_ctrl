package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trailblazer's Boots — Artifact — Equipment, {2}:
//
//	"Equipped creature has nonbasic landwalk. (It can't be blocked as
//	 long as defending player controls a nonbasic land.)
//	 Equip {2}"
//
// The reference card for nonbasic landwalk (#705), the one landwalk
// that names a SUPERTYPE rather than a land type: it bites on any land
// without the basic supertype (CR 205.4c), so a dual land or a utility
// land switches it on and a basic Island does not. The grant is layer
// 6 through the same GrantToAttached Swiftfoot Boots uses, and the
// block check reads it in game/landwalk.go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "634d5009-cbf3-44cb-8c15-7057f501a210",
		Name:         "Trailblazer's Boots",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.GrantToAttached("nonbasic landwalk"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}
