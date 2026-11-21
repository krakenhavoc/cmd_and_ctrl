package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sage's Nouliths — Artifact — Equipment {1}{U}:
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature gets +1/+0, has 'Whenever this creature attacks,
//	 untap target attacking creature,' and is a Cleric in addition to
//	 its other types.
//	 Hagneia — Equip {3}"
//
// The attack trigger is a granted ability (ADR 0093): it belongs to the
// equipped creature, goes away with the Equipment, and has the creature
// as its source. The target is chosen as the trigger goes on the stack
// and may be the equipped creature itself.
//
// No simplification.
func init() {
	const untapKey = "sages-nouliths/untap-attacker"
	untap := effects.Targeting(
		effects.On(game.EventAttack, effects.Self, "Sage's Nouliths — untap target attacking creature",
			effects.UntapTheTarget),
		effects.TargetCreature("target attacking creature", effects.AttackingCreature()))
	effects.Register(effects.Spec{
		OracleID:     "97248734-6e99-42a2-8f11-f3337b25f8ab",
		Name:         "Sage's Nouliths",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       untapKey,
			Triggered: []game.TriggeredAbility{untap},
			Text:      "Whenever this creature attacks, untap target attacking creature.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantAbilitiesToAttached(untapKey),
			effects.AttachedIsAlso("Cleric"),
		},
		Triggered: []game.TriggeredAbility{effects.JobSelect("Sage's Nouliths")},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
