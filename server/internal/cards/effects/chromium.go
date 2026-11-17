package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chromium — Legendary Creature — Elder Dragon {2}{W}{W}{U}{U}{B}{B}, 7/7
// (Legends, #5769):
//
//	"Flying
//	 Rampage 2 (Whenever this creature becomes blocked, it gets
//	 +2/+2 until end of turn for each creature blocking it beyond the
//	 first.)
//	 At the beginning of your upkeep, sacrifice Chromium unless you
//	 pay {W}{U}{B}."
//
// Flying is the engine's keyword and rampage its keyword trigger (CR
// 702.23, game/rampage.go). The upkeep payment is
// UpkeepSacrificeUnlessYouPay, the Legends Elder Dragons' shared shape:
// pay {W}{U}{B} or sacrifice it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7c91ae5d-0320-46a7-98d2-df0918202478",
		Name:            "Chromium",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "rampage 2"},
		Triggered: []game.TriggeredAbility{
			UpkeepSacrificeUnlessYouPay("Chromium", "{W}{U}{B}"),
		},
	})
}
