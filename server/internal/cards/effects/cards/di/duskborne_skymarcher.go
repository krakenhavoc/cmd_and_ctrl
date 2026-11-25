package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Duskborne Skymarcher — Creature — Vampire Cleric {W}, 1/1 (Ixalan):
//
//	"Flying
//	 {W}, {T}: Target attacking Vampire gets +1/+1 until end of turn."
//
// The target is an attacking Vampire (any player's, the Skymarcher itself if it
// is attacking), checked as the ability goes on the stack and again as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f7ff0046-8365-4ea4-9872-583d140f7b3f",
		Name:            "Duskborne Skymarcher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Target attacking Vampire gets +1/+1 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target attacking Vampire", effects.AttackingCreature(), effects.Subtype("Vampire")),
			Effect:  effects.BoostTargetUntilEOT("Duskborne Skymarcher — +1/+1 until end of turn", 1, 1),
		}},
	})
}
