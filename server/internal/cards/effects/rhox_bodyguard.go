package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rhox Bodyguard — Creature — Rhino Monk Soldier {3}{G}{W}, 2/3 (Conflux,
// #5587):
//
//	"Exalted (Whenever a creature you control attacks alone, that
//	 creature gets +1/+1 until end of turn.)
//	 When this creature enters, you gain 3 life."
//
// Exalted is the engine's keyword (game/exalted.go, CR 702.83a). The
// enters trigger uses the stack, so it can be answered, and gains the
// life as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "34bd39ed-4058-4c65-ba5f-9ac5cf7c166a",
		Name:            "Rhox Bodyguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
		Purpose:         game.Purpose{LifeGain: 3},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Rhox Bodyguard — gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
