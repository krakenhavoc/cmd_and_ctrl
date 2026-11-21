package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Basking Broodscale — Creature — Eldrazi Lizard {1}{G}, 2/2 (Modern Horizons 3):
//
//	"Devoid
//	 {1}{G}: Adapt 1.
//	 Whenever one or more +1/+1 counters are put on this creature, you may
//	 create a 0/1 colorless Eldrazi Spawn creature token with "Sacrifice this
//	 token: Add {C}.""
func init() {
	Register(Spec{
		OracleID:        "a7c69802-99c4-4778-934e-cc09ba58c267",
		Name:            "Basking Broodscale",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: Adapt 1.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{G}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return adaptN(g, item, 1) },
		}},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCounterPlaced, gapB51CountersPlacedOnThis,
				"Basking Broodscale — you may create a 0/1 colorless Eldrazi Spawn creature token", gapB51EldraziSpawn),
				"Basking Broodscale — create an Eldrazi Spawn token?"),
		},
	})
}
