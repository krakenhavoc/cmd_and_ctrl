package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fallen Ferromancer — Creature — Phyrexian Human Shaman {3}{R}, 1/1 (New Phyrexia, #5575):
//
//	"Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
//	 {1}{R}, {T}: This creature deals 1 damage to any target."
//
// Infect (CR 702.90) and Anaba Shaman's ping for {1}{R}. The ping is
// dealt by a source with infect, so a creature is dealt it as -1/-1
// counters and a player as poison counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "22446354-e5c3-4f19-9731-79a4562639d1",
		Name:            "Fallen Ferromancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, {T}: This creature deals 1 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
