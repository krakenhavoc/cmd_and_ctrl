package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tuinvale Guide — Creature — Faerie Scout {3}{W}, 2/3 (Wilds of
// Eldraine, #4471):
//
//	"Flying
//	 Celebration — This creature gets +1/+0 and has lifelink as long as
//	 two or more nonland permanents entered the battlefield under your
//	 control this turn."
//
// Armory Mice's layer 7c pump at +1/+0 and Gallant Pie-Wielder's layer 6
// grant of lifelink, both on the turn's entry tally (celebration.go): the
// Guide counts itself on the turn it enters, and both end with the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dab16c93-9788-46e0-9345-838ebeda8d54",
		Name:            "Tuinvale Guide",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.SelfPumpWhileCelebration(1, 0),
			effects.SelfKeywordWhileCelebration("lifelink"),
		},
	})
}
