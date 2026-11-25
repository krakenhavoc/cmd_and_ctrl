package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sylvia Brightspear — Legendary Creature — Human Knight {2}{W}, 2/2 (Battlebond, #5529):
//
//	"Partner with Khorvath Brightflame (When this creature enters, target player may put Khorvath into their hand from their library, then shuffle.)"
//	"Double strike"
//	"Dragons your team controls have double strike."
//
// Partner with is effects.PartnerWith (CR 702.124j). "Your team" is you: the table has no teams, as for Rushblade Commander.
//
// No simplification.
func init() {
	dragons := effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Dragon"}, YoursOnly: true}, "double strike")
	dragons.Label = "Dragons your team controls have double strike."
	effects.Register(effects.Spec{
		OracleID:        "22a93a6d-fe58-4135-8ac2-6b92e9bb75a3",
		Name:            "Sylvia Brightspear",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Static:          []game.StaticAbility{dragons},
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Sylvia Brightspear", "Khorvath Brightflame"),
		},
	})
}
