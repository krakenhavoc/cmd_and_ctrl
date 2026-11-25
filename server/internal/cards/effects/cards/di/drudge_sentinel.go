package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drudge Sentinel — Creature — Skeleton Warrior {2}{B}, 2/1 (Dominaria, S82
// OLD, #5530):
//
//	"{3}: Tap this creature. It gains indestructible until end of turn."
//
// Popular Egotist's tap-and-protect body: both halves belong to the resolving
// ability, so it grants indestructible even if the creature was already
// tapped. No tap in the cost, so it can be activated while tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "71ecb454-3598-4860-bba1-b3932f906171",
		Name:         "Drudge Sentinel",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Tap this creature. It gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{3}"),
			Effect:  effects.B23IndestructibleAndTapThis("Drudge Sentinel — gains indestructible until end of turn"),
		}},
	})
}
