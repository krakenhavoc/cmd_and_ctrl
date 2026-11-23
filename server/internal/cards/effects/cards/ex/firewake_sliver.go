package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b5a3c09d-e822-4e26-b569-7f5830a39e6a",
		Name:         "Firewake Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: firewakeSliverGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{1}, Sacrifice this permanent: Target Sliver creature gets +2/+2 until end of turn",
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
				Targets: effects.TargetCreature("target Sliver creature", effects.HasSubtype("Sliver")),
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
				Effect:  effects.BoostTargetUntilEOT("Firewake Sliver — +2/+2 until end of turn", 2, 2),
			}},
			Text: "{1}, Sacrifice this permanent: Target Sliver creature gets +2/+2 until end of turn.",
		}},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, "haste"),
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, firewakeSliverGrant),
		},
	})
}
