package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Armory Mice — Creature — Mouse {1}{W}, 3/1 (Wilds of Eldraine, #5164):
//
//	"Celebration — This creature gets +0/+2 as long as two or more
//	 nonland permanents entered the battlefield under your control this
//	 turn."
//
// A layer 7c static on the turn's entry tally (celebration.go): the Mice
// count themselves on the turn they enter, an artifact creature counts
// once, a token counts and a land does not, and one that has since left
// still counts. The bonus ends with the turn, when the tally resets.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "68f7d1bf-48c7-483e-b17f-a60008854b46",
		Name:         "Armory Mice",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{SelfPumpWhileCelebration(0, 2)},
	})
}
