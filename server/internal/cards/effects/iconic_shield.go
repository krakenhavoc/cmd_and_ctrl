package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	attack := WheneverThisAttacks(label, firstTargetGainsUntilEndOfTurn(0, label, "indestructible"))
	attack.Targets = Another(TargetCreature("another target attacking creature", AttackingCreature()))
	Register(Spec{
		OracleID:     "1c4fdd52-dc99-4942-bdaa-93f74dd4ebdc",
		Name:         "Iconic Shield",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       grantKey,
			Triggered: []game.TriggeredAbility{attack},
			Text:      "Whenever this creature attacks, another target attacking creature gains indestructible until end of turn.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(1, 2),
			GrantAbilitiesToAttached(grantKey),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}
