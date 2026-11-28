package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kjeldoran Elite Guard — Creature — Human Soldier {3}{W}, 2/2 (Ice Age, S82
// OLD, #5636):
//
//	"{T}: Target creature gets +2/+2 until end of turn. When that creature leaves
//	 the battlefield this turn, sacrifice this creature. Activate only during
//	 combat."
//
// Runesword's delayed leaves-the-battlefield trigger on the target, sacrificing
// the source; any step from the beginning of combat to the end of combat.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1f6df2a-fc97-44d4-a0d1-902d8df94002",
		Name:         "Kjeldoran Elite Guard",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Target creature gets +2/+2 until end of turn. When that creature leaves the battlefield this turn, sacrifice this creature. Activate only during combat.",
			Purpose:   effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
			Cost:      effects.TapCost(),
			Targets:   effects.TargetCreature("target creature"),
			Condition: effects.OldB835DuringCombat(),
			Effect:    effects.OldB835TargetGetsThenSacrificeThisWhenItLeaves(2, 2, "Kjeldoran Elite Guard — +2/+2"),
		}},
	})
}
