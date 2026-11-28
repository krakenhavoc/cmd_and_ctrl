package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Homarid Warrior — Creature — Homarid Warrior {4}{U}, 3/3 (Fallen Empires, S82
// OLD batch 852, #5637):
//
//	"{U}: This creature gains shroud until end of turn and doesn't untap during
//	 your next untap step. Tap it. (A creature with shroud can't be the target of
//	 spells or abilities.)"
//
// The three instructions in printed order (S82OldB852ShroudFreezeAndTap, shared
// with Deep Spawn): shroud until end of turn, the freeze on this object, then the
// tap. No {T} cost, so it may be activated while summoning sick.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6facb0f6-fa49-4ce6-9f95-15bb3ce4bcf2",
		Name:         "Homarid Warrior",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: This creature gains shroud until end of turn and doesn't untap during your next untap step. Tap it.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{U}"),
			Effect:  effects.S82OldB852ShroudFreezeAndTap("Homarid Warrior"),
		}},
	})
}
