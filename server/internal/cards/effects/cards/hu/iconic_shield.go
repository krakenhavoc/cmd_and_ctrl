package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iconic Shield — {W} Artifact — Equipment (Marvel Super Heroes Commander,
// #4455):
//
//	"Equipped creature gets +1/+2 and has "Whenever this creature attacks,
//	 another target attacking creature gains indestructible until end of
//	 turn."
//	 Equip {3}"
//
// The attack trigger is a granted ability (ADR 0093) with Helitrooper's
// target and body, naming indestructible.
//
// No simplification.
func init() {
	const grantKey = "iconic-shield/indestructible"
	const label = "Iconic Shield — another attacking creature gains indestructible until end of turn"
	attack := effects.WheneverThisAttacks(label, effects.FirstTargetGainsUntilEndOfTurn(0, label, "indestructible"))
	attack.Targets = effects.Another(effects.TargetCreature("another target attacking creature", effects.AttackingCreature()))
	effects.Register(effects.Spec{
		OracleID:     "1c4fdd52-dc99-4942-bdaa-93f74dd4ebdc",
		Name:         "Iconic Shield",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       grantKey,
			Triggered: []game.TriggeredAbility{attack},
			Text:      "Whenever this creature attacks, another target attacking creature gains indestructible until end of turn.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 2),
			effects.GrantAbilitiesToAttached(grantKey),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{3}"),
		},
	})
}
