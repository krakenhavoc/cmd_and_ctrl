package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rootha, Mercurial Artist — {1}{U}{R} Legendary Creature — Orc Shaman:
//
//	"{2}, Return Rootha to its owner's hand: Copy target instant or
//	 sorcery spell you control. You may choose new targets for the
//	 copy."
//
// #2028: the return is the cost (ReturnThis), paid at announce
// (CR 602.2b). The copy is Lithoform Engine's (copyTargetedSpell),
// with its "you may choose new targets" prompt. As a commander, its
// owner is asked whether Rootha goes to the command zone instead
// (CR 903.9b) before anything is paid.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3735559f-efe5-43ad-9b3e-3ef127c0ceda",
		Name:         "Rootha, Mercurial Artist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Return Rootha to its owner's hand: Copy target instant or sorcery spell you control. You may choose new targets for the copy.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.ReturnThis()),
			Targets: effects.TargetInstantOrSorcerySpell("target instant or sorcery spell you control", effects.YouControl()),
			Effect:  effects.CopyTargetedSpell,
		}},
	})
}
