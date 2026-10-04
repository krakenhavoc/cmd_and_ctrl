package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kitchen Finks — Creature — Ouphe {1}{G/W}{G/W}, 3/2:
//
//	"When this creature enters, you gain 2 life.
//	 Persist"
//
// Persist is PrintedKeywords (game/undying_persist.go, #2075), so a Finks
// that dies with no -1/-1 counter returns with one and gains 2 life again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5470dcfa-4eff-43da-abf7-19922841f719",
		Name:            "Kitchen Finks",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordPersist},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kitchen Finks — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
