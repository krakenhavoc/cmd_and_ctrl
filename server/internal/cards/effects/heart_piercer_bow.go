package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heart-Piercer Bow — Artifact — Equipment {2} (Commander Masters):
//
//	"Whenever equipped creature attacks, this Equipment deals 1 damage to
//	 target creature defending player controls.
//	 Equip {1}"
//
// Thunder Lasso's shape: the Equipment's own trigger on the attached
// creature's attack, its target clause reading the defending player off the
// attacker (mshB2TargetCreatureTheAttackerDefends), with Blasting Station's
// 1 damage from the source. With no creature to target the trigger is
// removed (CR 603.3d); a target that is gone at resolution does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ebe9dad7-d74d-4fa8-b17a-88ff79209379",
		Name:         "Heart-Piercer Bow",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attachedCreatureAttacked(ev, source)
			},
			TargetsFrom: mshB2TargetCreatureTheAttackerDefends,
			Key:         "Heart-Piercer Bow — 1 damage to target creature defending player controls",
			Effect:      b33DamageChosenTargetFromSource(1),
		}},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}
