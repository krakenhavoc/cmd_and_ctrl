package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meteoric Mace — Artifact — Equipment {4}{R}{R} (Modern Horizons 3, #5948):
//
//	"Equipped creature gets +4/+0 and has trample.
//	 Equip {4}
//	 Cascade (When you cast this spell, exile cards from the top of your
//	 library until you exile a nonland card that costs less. You may cast it
//	 without paying its mana cost. Put the exiled cards on the bottom in a
//	 random order.)"
//
// Cascade is the engine keyword whole (Annoyed Altisaur's): the trigger goes on
// the stack above the Mace when it is cast, so the free spell resolves first,
// and it compares against the Mace's mana value, six. The rest is the shared
// +4/+0 pump, a layer-6 trample grant to the equipped creature, and equip {4}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6a607f2-4228-4bcc-84ed-37bd362d7542",
		Name:         "Meteoric Mace",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Cascade()},
		Static: []game.StaticAbility{
			effects.PumpAttached(4, 0),
			effects.GrantToAttached("trample"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
