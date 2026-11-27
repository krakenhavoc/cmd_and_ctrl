package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aerial Caravan — Creature — Human Soldier {4}{U}{U}, 4/3 (Mercadian
// Masques, S82 OLD b682, #5619):
//
//	"Flying
//	 {1}{U}{U}: Exile the top card of your library. Until end of turn, you may
//	 play that card. (Reveal the card as you exile it.)"
//
// Flying is a printed keyword. The impulse is Abbot of Keral Keep's
// (ExileTopWithPermission): the card goes to exile face up and its controller
// may play it, a land included, until end of turn. It may be activated as
// often as it can be paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d13aff7c-f0ee-44e2-ae1c-62cd86ad46a6",
		Name:            "Aerial Caravan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}{U}: Exile the top card of your library. Until end of turn, you may play that card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{1}{U}{U}"),
			Effect:  effects.OutpostSiegeImpulse,
		}},
	})
}
