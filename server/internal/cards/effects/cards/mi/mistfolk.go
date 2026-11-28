package cardsmi

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Mistfolk — Creature — Illusion {U}{U}, 1/2 (Ice Age, S82 OLD batch 850, #5636):
//
//	"{U}: Counter target spell that targets this creature."
//
// Daring Apprentice's counter over a spell clause judged against the ability's
// own source (game.TargetSpec.ItemTargetsSource, effects.TargetsThisCreature,
// the looser sibling of Silver Wyvern's "targets only"). It is checked as the
// ability is activated and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d57409dc-22fa-40c1-a043-cab94d5f3b5f",
		Name:         "Mistfolk",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: Counter target spell that targets this creature.",
			Cost:    effects.ManaCost("{U}"),
			Targets: effects.TargetsThisCreature(),
			Effect:  effects.CounterTheChosenSpell,
		}},
	})
}
