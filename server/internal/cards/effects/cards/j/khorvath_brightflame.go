package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Khorvath Brightflame — Legendary Creature — Dragon {5}{R}, 3/4 (Battlebond, #5529):
//
//	"Partner with Sylvia Brightspear (When this creature enters, target player may put Sylvia into their hand from their library, then shuffle.)"
//	"Flying, haste"
//	"Knights your team controls have flying and haste."
//
// Partner with is effects.PartnerWith; the grant is Rushblade Commander's, once per keyword. "Your team" is you.
//
// No simplification.
func init() {
	knights := effects.TribeFilter{Tribes: []string{"Knight"}, YoursOnly: true}
	flying := effects.TribalKeywordGrant(knights, "flying")
	flying.Label = "Knights your team controls have flying."
	haste := effects.TribalKeywordGrant(knights, "haste")
	haste.Label = "Knights your team controls have haste."
	effects.Register(effects.Spec{
		OracleID:        "b96b1341-0c15-4816-be2d-5cc98b2afb63",
		Name:            "Khorvath Brightflame",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Static:          []game.StaticAbility{flying, haste},
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Khorvath Brightflame", "Sylvia Brightspear"),
		},
	})
}
