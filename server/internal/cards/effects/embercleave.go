package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Embercleave — Legendary Artifact — Equipment {4}{R}{R}:
//
//	"Flash
//	 This spell costs {1} less to cast for each attacking creature you
//	 control.
//	 When Embercleave enters, attach it to target creature you
//	 control.
//	 Equipped creature gets +1/+1 and has double strike and trample.
//	 Equip {3}"
//
// Four printed clauses, four pieces already in the catalog:
//
//   - Flash is a printed keyword.
//   - The cost reduction is Ancient Stone Idol's CostsLessEach shape,
//     narrowed to "you control" with PermanentsYouControl —
//     AttackingCreature() already reads live combat state
//     (c.AttackingTarget != uuid.Nil).
//   - "When ~ enters, attach it to target creature you control" is
//     Maul of the Skyclaves' ETB attach, verbatim: Targeting wraps
//     WhenThisEnters, and the resolution is the same
//     AttachSourceToTarget the equip ability itself uses, so a
//     creature that dies in response leaves Embercleave unattached
//     rather than erroring.
//   - The static grant is PumpAttached + GrantToAttached, as usual.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4d6120d6-fcce-40bc-9fc6-e1f5beb6c728",
		Name:            "Embercleave",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(PermanentsYouControl(AttackingCreature()),
				"This spell costs {1} less to cast for each attacking creature you control."),
		},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("double strike", "trample"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Embercleave — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}
