package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Writhing Chrysalis — Creature — Eldrazi Drone {2}{R}{G}, 2/3 (Modern Horizons 3):
//
//	"Devoid
//	 When you cast this spell, create two 0/1 colorless Eldrazi Spawn creature
//	 tokens with "Sacrifice this token: Add {C}."
//	 Reach
//	 Whenever you sacrifice another Eldrazi, put a +1/+1 counter on this
//	 creature."
func init() {
	Register(Spec{
		OracleID:        "d27bebc9-348d-4b8e-816a-a4da66fb5618",
		Name:            "Writhing Chrysalis",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Tokens: 2, CreatureTokens: 2},
		PrintedKeywords: []string{game.KeywordDevoid, "reach"},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Writhing Chrysalis — create two 0/1 colorless Eldrazi Spawn creature tokens", gapB52EldraziSpawn(2)),
			On(game.EventSacrifice, gapB52SacrificedAnotherEldrazi,
				"Writhing Chrysalis — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
