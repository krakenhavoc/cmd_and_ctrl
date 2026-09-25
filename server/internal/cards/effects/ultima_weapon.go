package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ultima Weapon — Legendary Artifact — Equipment for {7}:
//
//	"Whenever equipped creature attacks, destroy target creature an
//	 opponent controls.
//	 Equipped creature gets +7/+7.
//	 Equip {7}"
//
// Argentum Armor's narrower, bigger sibling: a +7/+7 that removes a
// blocker every time it swings.
//
// The trigger is the EQUIPMENT's own, not one granted to the host,
// which is the distinction that keeps this card off the granted-
// ability seam Kaldra Compleat waits on. "Whenever equipped creature
// attacks" is a condition on an ability printed on the Equipment —
// attachedCreatureAttacked reads EventAttack's CardID against the
// attachment relation — so the harvester finds it on Ultima Weapon's
// own catalog entry, exactly where it looks.
//
// The target clause is narrower than Argentum Armor's "target
// permanent" in both directions of the sentence: creatures only, and
// only ones an OPPONENT controls. That second half means the trigger
// simply does not fire when no opponent has a creature (CR 603.3d,
// no legal target, no prompt) rather than forcing you to blow up
// your own board.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0dc4bf39-67d6-44e5-8993-dd18b20cfdee",
		Name:         "Ultima Weapon",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(7, 7)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			Key:     "Ultima Weapon — destroy target creature an opponent controls",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attachedCreatureAttacked(ev, source)
			},
			Targets: TargetCreature("target creature an opponent controls", OpponentControls()),
			Effect:  destroyChosenPermanent,
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{7}"),
		},
	})
}
