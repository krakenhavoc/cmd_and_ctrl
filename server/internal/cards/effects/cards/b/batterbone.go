package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Batterbone — Artifact — Equipment for {2}:
//
//	"Living weapon (When this Equipment enters, create a 0/0 black
//	 Phyrexian Germ creature token, then attach this to it.)
//	 Equipped creature gets +1/+1 and has vigilance and lifelink.
//	 Equip {5}"
//
// Batterskull's two-mana cousin: the same three keywords for a
// quarter of the cost and a much smaller body, and a turn-two 1/1
// lifelink that an aristocrats deck is happy to sacrifice.
//
// Nothing here is new. It is the shared LivingWeapon trigger, the
// same 7c pump and layer-6 grant every Equipment in the catalog
// uses, and one equip ability — which is the point of having built
// the keyword as a helper.
//
// The expensive equip cost relative to the cast cost is printed and
// load-bearing: moving Batterbone onto a real creature is meant to
// be the late-game plan, not the turn-three one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "389d459d-446a-4b84-82b2-a30fe6ced11f",
		Name:         "Batterbone",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LivingWeapon("Batterbone"),
		},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("vigilance", "lifelink"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{5}"),
		},
	})
}
