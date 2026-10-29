package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain America's Shield — {2} Legendary Artifact — Equipment (Marvel
// Super Heroes, #4455):
//
//	"Indestructible
//	 Equipped creature gets +0/+8 and has vigilance.
//	 Whenever equipped creature attacks, tap target creature defending
//	 player controls.
//	 Equip {2}"
//
// The attack trigger is Grappling Hook's shape (the attached creature's
// attack event) with a target clause that reads the defending player off
// the ATTACKER, not off the Equipment (mshB2TargetCreatureTheAttackerDefends):
// the player that creature is attacking, so a planeswalker or battle attack
// finds its protector or controller. The target is chosen as the trigger
// goes on the stack and tapped as it resolves; with no creature to target
// the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8255a797-ddd3-4db0-a3f5-dd8f38c3b908",
		Name:            "Captain America's Shield",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static: []game.StaticAbility{
			PumpAttached(0, 8),
			GrantToAttached("vigilance"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attachedCreatureAttacked(ev, source)
			},
			TargetsFrom: mshB2TargetCreatureTheAttackerDefends,
			Key:         "Captain America's Shield — tap target creature defending player controls",
			Effect:      b36TapChosenCreature,
		}},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
