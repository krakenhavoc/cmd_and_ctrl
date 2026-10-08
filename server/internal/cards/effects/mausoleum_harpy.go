package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mausoleum Harpy — Creature — Harpy {4}{B}, 3/3:
//
//	"Flying
//	 Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 Whenever another creature you control dies, if you have the city's
//	 blessing, put a +1/+1 counter on this creature."
//
// "Another" excludes the Harpy itself, and "you control" reads the dead
// creature's controller as it left (ACreatureYouControlDied). The "if"
// is intervening (CR 603.4) and is checked again at resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ba682da8-50f5-478f-99cc-60a57d229a04",
		Name:            "Mausoleum Harpy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordAscend},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, AllOf(AnotherCreatureDied, ACreatureYouControlDied, YouHaveTheCitysBlessingNow),
				"Mausoleum Harpy — put a +1/+1 counter on it", blessedSelfCounter),
		},
	})
}
