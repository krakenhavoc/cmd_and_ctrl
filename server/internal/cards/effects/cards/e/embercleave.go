package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "4d6120d6-fcce-40bc-9fc6-e1f5beb6c728",
		Name:            "Embercleave",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant}, // #4734: flash; +1/+1, double strike and trample on the attacker
		PrintedKeywords: []string{"flash"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.PermanentsYouControl(effects.AttackingCreature()),
				"This spell costs {1} less to cast for each attacking creature you control."),
		},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("double strike", "trample"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Embercleave — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{3}"),
		},
	})
}
