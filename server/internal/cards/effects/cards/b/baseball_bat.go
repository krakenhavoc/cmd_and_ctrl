package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baseball Bat — Artifact — Equipment {G}{W} (Duskmourn: House of Horror,
// #4466):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +1/+1.
//	 Whenever equipped creature attacks, tap up to one target creature.
//	 Equip {3}"
//
// The entry trigger is Celestial Armor's attach (AttachSourceToTarget): it
// targets on the stack and a target that has gone gets nothing (CR 608.2b).
// The pump is PumpAttached. The attack trigger is on the Equipment, keyed on
// the creature it is attached to (attachedCreatureAttacked, Grappling
// Hook's condition), with an "up to one target creature" clause: it may
// have no target, and a target that is gone or hexproofed by resolution is
// skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b5a3765-943b-4925-b008-fc734b2934f7",
		Name:         "Baseball Bat",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Baseball Bat — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
			effects.Targeting(
				effects.On(game.EventAttack,
					func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
						return effects.AttachedCreatureAttacked(ev, source)
					},
					"Baseball Bat — tap up to one target creature",
					effects.Msc13TapTheLegalTargets),
				effects.UpToOneTargetCreature("up to one target creature"),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
