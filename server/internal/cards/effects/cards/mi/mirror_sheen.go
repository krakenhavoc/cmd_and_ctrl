package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirror Sheen — {1}{U/R}{U/R} Enchantment:
//
//	"{1}{U/R}{U/R}: Copy target instant or sorcery spell that targets you. You
//	 may choose new targets for the copy."
//
// You control the copy, whoever cast the original.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6494704-54b2-4c80-acb4-184c9815c5f3",
		Name:         "Mirror Sheen",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U/R}{U/R}: Copy target instant or sorcery spell that targets you. You may choose new targets for the copy.",
			Cost:    effects.ManaCost("{1}{U/R}{U/R}"),
			Targets: effects.TargetInstantOrSorcerySpell("target instant or sorcery spell that targets you", effects.S82OldB523SpellTargetsYou()),
			Effect:  effects.CopyTargetedSpellWithNewTargets,
		}},
	})
}
