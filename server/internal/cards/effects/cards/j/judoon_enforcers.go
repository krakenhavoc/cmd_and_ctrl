package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Judoon Enforcers — Creature — Alien Rhino Soldier {5}{R}{W}, 8/8
// (EDHREC rank 12555):
//
//	"Trample
//	 No more than one creature can attack you each combat.
//	 Suspend 6—{1}{R}{W}"
//
// A Crawlspace with a bound of one on an 8/8 body (#1507). Trample is
// an engine keyword, and suspend is the keyword's own machinery end to
// end — the special action, the countdown, the free cast and the
// CR 702.62a haste (#659) — so the card file is three declarations.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ca04089c-24b6-465e-9303-ea28c0d6f3c7",
		Name:            "Judoon Enforcers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		AttackLimits:    []game.AttackLimit{effects.NoMoreThanNCanAttackYouEachCombat(1)},
		SpecialActions: []game.SpecialAction{
			effects.Suspend(6, "{1}{R}{W}"),
		},
	})
}
