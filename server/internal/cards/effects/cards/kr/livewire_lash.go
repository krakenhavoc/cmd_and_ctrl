package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Livewire Lash — Artifact — Equipment {2} (Double Masters 2022):
//
//	"Equipped creature gets +2/+0 and has "Whenever this creature becomes
//	 the target of a spell, this creature deals 2 damage to any target."
//	 Equip {2}"
//
// The trigger is a granted ability (ADR 0093): it belongs to the equipped
// creature, goes away with the Equipment and has the creature as its source,
// so the creature deals the damage. "Becomes the target of a spell" reads
// the stack item, so an ability that targets the creature does not trigger
// it (Bonecrusher Giant's SelfTargetedByASpell). The any-target is chosen
// as the trigger is put on the stack, above the spell.
//
// No simplification.
func init() {
	const grantKey = "livewire-lash/targeted-by-spell-ping"
	ping := effects.On(game.EventBecomesTarget, effects.SelfTargetedByASpell,
		"Livewire Lash — this creature deals 2 damage to any target", effects.DamageToFirstTarget(2))
	ping.Targets = effects.TargetAny()
	ping.Purpose = effects.ForTargets(effects.DamageToTarget(0, 2))
	effects.Register(effects.Spec{
		OracleID:     "9d3f1b93-ec3b-47eb-bd93-89890fd33ccb",
		Name:         "Livewire Lash",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       grantKey,
			Triggered: []game.TriggeredAbility{ping},
			Text:      "Whenever this creature becomes the target of a spell, this creature deals 2 damage to any target.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			effects.GrantAbilitiesToAttached(grantKey),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
