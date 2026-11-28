package cardsbl

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Brine Shaman — Creature — Human Cleric Shaman {1}{B}, 1/1 (Ice Age, S82 OLD
// batch 850, #5636):
//
//	"{T}, Sacrifice a creature: Target creature gets +2/+2 until end of turn.
//	 {1}{U}{U}, Sacrifice a creature: Counter target creature spell."
//
// The pump is Blighted Shaman's second ability; the counter is Daring
// Apprentice's with a mana-and-sacrifice cost, aimed at a creature spell. The
// creature sacrificed may be the Shaman itself or the pump's target (the pump
// then does nothing, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e3e2f71-2159-4f83-a9c1-a67ecac8a711",
		Name:         "Brine Shaman",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}, Sacrifice a creature: Target creature gets +2/+2 until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
				Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "a creature", effects.Creature())),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BoostTargetUntilEOT("Brine Shaman — +2/+2", 2, 2),
			},
			{
				Label:   "{1}{U}{U}, Sacrifice a creature: Counter target creature spell.",
				Cost:    effects.Plus(effects.ManaCost("{1}{U}{U}"), effects.SacrificeN(1, "a creature", effects.Creature())),
				Targets: effects.TargetSpell("target creature spell", effects.Creature()),
				Effect:  effects.CounterTheChosenSpell,
			},
		},
	})
}
