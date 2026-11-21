package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunder Lasso — Artifact — Equipment {2}{W} (Outlaws of Thunder
// Junction, #4468):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +1/+1.
//	 Whenever equipped creature attacks, tap target creature defending
//	 player controls.
//	 Equip {2}"
//
// The entry attach is Falcon's Wing Harness's, and the attack trigger is
// Captain America's Shield's: the Equipment's own trigger on the attached
// creature's attack, its target clause reading the defending player off
// the attacker (mshB2TargetCreatureTheAttackerDefends). With no creature
// to target the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d98cc18b-6ebb-4f55-9d0d-b8fbe9a92bc1",
		Name:         "Thunder Lasso",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Thunder Lasso — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttachedCreatureAttacked(ev, source)
				},
				TargetsFrom: effects.MshB2TargetCreatureTheAttackerDefends,
				Key:         "Thunder Lasso — tap target creature defending player controls",
				Effect:      effects.B36TapChosenCreature,
			},
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
