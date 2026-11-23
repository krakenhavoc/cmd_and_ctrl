package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// High-Flying Ace — {2}{W} Creature — Bird Mutant 2/3 (Teenage Mutant
// Ninja Turtles, #4457):
//
//	"Flying
//	 {3}{W}: Target creature without flying gains flying until end of
//	 turn. Activate only as a sorcery."
//
// Alabaster Mage's grant at sorcery speed (CR 602.5d). "Without flying"
// is read on the target's effective abilities when it is chosen and
// again as the ability resolves (CR 608.2b), so a creature that gained
// flying in response is no longer a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "84a56740-8563-4ce5-8560-d44998d7f80d",
		Name:            "High-Flying Ace",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{W}: Target creature without flying gains flying until end of turn. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:         effects.ManaCost("{3}{W}"),
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature without flying", effects.WithoutKeyword("flying")),
			Effect:       effects.JmpB7FirstLegalTargetGains("High-Flying Ace — flying until end of turn", "flying"),
		}},
	})
}
