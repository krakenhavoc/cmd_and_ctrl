package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Herald of Leshrac — Creature — Avatar {6}{B}, 2/4 (Coldsnap, S82 OLD b559):
//
//	"Flying
//	 Cumulative upkeep—Gain control of a land you don't control.
//	 This creature gets +1/+1 for each land you control but don't own.
//	 When this creature leaves the battlefield, each player gains control of
//	 each land they own that you control."
//
// One land per age counter, each named by the payer and taken for as long as
// they have it (CR 702.24a); the bonus counts the lands the controller has but
// does not own, however they were taken; the leave trigger hands every such
// land to its owner.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ed9e39ad-996c-4a68-8d11-e1c7d35f556c",
		Name:            "Herald of Leshrac",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.OldB559PumpThisPerLandYouDontOwn()},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeepPaying("Herald of Leshrac — cumulative upkeep: gain control of a land you don't control",
				effects.GainControlOfLandPayment(1, "land", "lands", effects.QueryTypes("land")), "land", "lands"),
			effects.WhenThisLeaves("Herald of Leshrac — each player gains control of each land they own that you control",
				effects.OldB559EachPlayerGetsTheirLandsBack),
		},
	})
}
