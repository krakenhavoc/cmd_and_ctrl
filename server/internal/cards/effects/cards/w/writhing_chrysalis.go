package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Writhing Chrysalis — Creature — Eldrazi Drone {2}{R}{G}, 2/3 (Modern Horizons 3):
//
//	"Devoid
//	 When you cast this spell, create two 0/1 colorless Eldrazi Spawn creature
//	 tokens with "Sacrifice this token: Add {C}."
//	 Reach
//	 Whenever you sacrifice another Eldrazi, put a +1/+1 counter on this
//	 creature."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d27bebc9-348d-4b8e-816a-a4da66fb5618",
		Name:            "Writhing Chrysalis",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Tokens: 2, CreatureTokens: 2},
		PrintedKeywords: []string{game.KeywordDevoid, "reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouCastThisSpell("Writhing Chrysalis — create two 0/1 colorless Eldrazi Spawn creature tokens", effects.GapB52EldraziSpawn(2)),
			effects.On(game.EventSacrifice, effects.GapB52SacrificedAnotherEldrazi,
				"Writhing Chrysalis — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
