package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Firewake Sliver — Creature — Sliver {1}{R}{G}, 1/1 (Time Spiral, #5596):
//
//	"All Sliver creatures have haste.
//	 All Slivers have "{1}, Sacrifice this permanent: Target Sliver
//	 creature gets +2/+2 until end of turn.""
//
// Two grants to every Sliver creature at the table: haste in layer 6 (CR
// 613.1f), and an activated ability (ADR 0093) whose cost sacrifices the
// Sliver that has it. Only that Sliver's controller can activate it (CR
// 602.2), and the target may be any Sliver creature, an opponent's
// included.
//
// No simplification.
const firewakeSliverGrant = "firewake-sliver/pump"

func init() {
	Register(Spec{
		OracleID:     "b5a3c09d-e822-4e26-b569-7f5830a39e6a",
		Name:         "Firewake Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: firewakeSliverGrant,
			Activated: []ActivatedAbility{{
				Label:   "{1}, Sacrifice this permanent: Target Sliver creature gets +2/+2 until end of turn",
				Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
				Targets: TargetCreature("target Sliver creature", HasSubtype("Sliver")),
				Purpose: ForTargets(BenefitUntilEOT(0, 2, 2)),
				Effect:  boostTargetUntilEOT("Firewake Sliver — +2/+2 until end of turn", 2, 2),
			}},
			Text: "{1}, Sacrifice this permanent: Target Sliver creature gets +2/+2 until end of turn.",
		}},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}}, "haste"),
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}}, firewakeSliverGrant),
		},
	})
}
